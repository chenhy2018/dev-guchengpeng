package m3u8

import (
	"bytes"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	xlog "github.com/qiniu/xlog.v1"
)

// MediaType for EXT-X-PLAYLIST-TYPE tag
type MediaType uint

const (
	// use 0 for undefined types
	// EVENT type
	EVENT MediaType = iota + 1
	// VOD type
	VOD
)

const (
	minver   = uint8(3)
	DATETIME = time.RFC3339Nano // Format for EXT-X-PROGRAM-DATE-TIME
)

var (
	// ErrPlaylistFull playlist is full
	ErrPlaylistFull = errors.New("playlist is full")
)

type MediaSegment struct {
	Title           string // optional second parameter for EXTINF tag
	URI             string
	Duration        float64   // first parameter for EXTINF tag; duration must be integers if protocol version is less than 3 but we are always keep them float
	ProgramDateTime time.Time // EXT-X-PROGRAM-DATE-TIME tag associates the first sample of a media segment with an absolute date and/or time
	IsDiscontinuity bool
}

type MediaPlaylist struct {
	Segments       []*MediaSegment
	MediaType      MediaType
	buf            bytes.Buffer
	Args           string // optional arguments placed after URIs (URI?Args)
	TargetDuration float64
	Winsize        uint   // max number of segments displayed in an encoded playlist; need set to zero for VOD playlists
	capacity       uint   // total capacity of slice used for the playlist
	head           uint   // head of FIFO, we add segments to head
	tail           uint   // tail of FIFO, we remove segments from tail
	count          uint   // number of segments added to the playlist
	SeqNo          uint64 // EXT-X-MEDIA-SEQUENCE
	Ver            uint8
	Iframe         bool // EXT-X-I-FRAMES-ONLY
	Closed         bool // is this VOD (closed) or Live (sliding) playlist?
	durationAsInt  bool // output durations as integers of floats?
	lastEndTime    int64
}

func (p *MediaPlaylist) Init(winsize uint, capacity uint) error {
	p.Ver = minver
	p.capacity = capacity
	if err := p.SetWinSize(winsize); err != nil {
		return err
	}
	p.Segments = make([]*MediaSegment, capacity)
	return nil
}

func strver(ver uint8) string {
	return strconv.FormatUint(uint64(ver), 10)
}

// SetWinSize overwrites the playlist's window size.
func (p *MediaPlaylist) SetWinSize(winsize uint) error {
	if winsize > p.capacity {
		return errors.New("capacity must be greater than winsize or equal")
	}
	p.Winsize = winsize
	return nil
}

// Append general chunk to the tail of chunk slice for a media playlist.
// This operation does reset playlist cache.
func (p *MediaPlaylist) AppendSegment(uri string, duration float64, title string, isVod bool) error {
	seg := new(MediaSegment)
	seg.URI = uri
	seg.Duration = duration
	seg.Title = title

	eles := strings.Split(uri, "/")
	offset := 3
	if isVod != true {
		offset = 5
	}
	startTime, err := strconv.ParseInt(eles[offset], 10, 64)
	if err != nil {
		return err
	}
	endTime, err := strconv.ParseInt(eles[offset+1], 10, 64)
	if err != nil {
		return err
	}

	if p.lastEndTime != -1 {
		if startTime-p.lastEndTime > 500 {
			seg.IsDiscontinuity = true
		}
	}
	p.lastEndTime = endTime

	return p.Append(seg)
}

// AppendSegment appends a MediaSegment to the tail of chunk slice for a media playlist.
// This operation does reset playlist cache.
func (p *MediaPlaylist) Append(seg *MediaSegment) error {
	if p.head == p.tail && p.count > 0 {
		return ErrPlaylistFull
	}
	p.Segments[p.tail] = seg
	p.tail = (p.tail + 1) % p.capacity
	p.count++
	if p.TargetDuration < seg.Duration {
		p.TargetDuration = math.Ceil(seg.Duration)
	}
	p.buf.Reset()
	return nil
}

func (p *MediaPlaylist) addVersion() {
	p.buf.WriteString("#EXTM3U\n#EXT-X-VERSION:")
	p.buf.WriteString(strver(p.Ver))
	p.buf.WriteRune('\n')
}

func (p *MediaPlaylist) addCacheInfo() {
	p.buf.WriteString("#EXT-X-ALLOW-CACHE:YES")
	p.buf.WriteRune('\n')
}

func (p *MediaPlaylist) addPlaylistType() {
	if p.MediaType > 0 {
		p.buf.WriteString("#EXT-X-PLAYLIST-TYPE:")
		switch p.MediaType {
		case EVENT:
			p.buf.WriteString("EVENT\n")
			p.buf.WriteString("#EXT-X-ALLOW-CACHE:NO\n")
		case VOD:
			p.buf.WriteString("VOD\n")
		}
	}
}

func (p *MediaPlaylist) addMediaSequence(sequence uint64) {
	p.buf.WriteString("#EXT-X-MEDIA-SEQUENCE:")
	p.buf.WriteString(strconv.FormatUint(sequence, 10))
	p.buf.WriteRune('\n')

	p.buf.WriteString("#EXT-X-DISCONTINUITY")
	p.buf.WriteRune('\n')
}

func (p *MediaPlaylist) addTargetDuraion() {
	p.buf.WriteString("#EXT-X-TARGETDURATION:")
	p.buf.WriteString(strconv.FormatInt(int64(math.Ceil(p.TargetDuration)), 10)) // due section 3.4.2 of M3U8 specs EXT-X-TARGETDURATION must be integer
	p.buf.WriteRune('\n')
}

func (p *MediaPlaylist) addIframe() {
	if p.Iframe {
		p.buf.WriteString("#EXT-X-I-FRAMES-ONLY\n")
	}
}

func (p *MediaPlaylist) addSegments() {
	var (
		seg           *MediaSegment
		durationCache = make(map[float64]string)
	)

	head := p.head
	count := p.count
	for i := uint(0); (i < p.Winsize || p.Winsize == 0) && count > 0; count-- {
		seg = p.Segments[head]
		head = (head + 1) % p.capacity
		if seg == nil { // protection from badly filled chunklists
			continue
		}
		if p.Winsize > 0 { // skip for VOD playlists, where winsize = 0
			i++
		}

		if !seg.ProgramDateTime.IsZero() {
			p.buf.WriteString("#EXT-X-PROGRAM-DATE-TIME:")
			p.buf.WriteString(seg.ProgramDateTime.Format(DATETIME))
			p.buf.WriteRune('\n')
		}
		if seg.IsDiscontinuity {
			p.buf.WriteString("#EXT-X-DISCONTINUITY\n")
		}
		p.buf.WriteString("#EXTINF:")
		if str, ok := durationCache[seg.Duration]; ok {
			p.buf.WriteString(str)
		} else {
			if p.durationAsInt {
				// Old Android players has problems with non integer Duration.
				durationCache[seg.Duration] = strconv.FormatInt(int64(math.Ceil(seg.Duration)), 10)
			} else {
				// Wowza Mediaserver and some others prefer floats.
				durationCache[seg.Duration] = strconv.FormatFloat(seg.Duration, 'f', 3, 32)
			}
			p.buf.WriteString(durationCache[seg.Duration])
		}
		p.buf.WriteRune(',')
		p.buf.WriteString(seg.Title)
		p.buf.WriteRune('\n')
		p.buf.WriteString(seg.URI)
		if p.Args != "" {
			p.buf.WriteRune('?')
			p.buf.WriteString(p.Args)
		}
		p.buf.WriteRune('\n')
	}
}

func (p *MediaPlaylist) addEndlist() {
	if p.Closed {
		p.buf.WriteString("#EXT-X-ENDLIST\n")
	}
}

// Encode Generate output in M3U8 format. Marshal `winsize` elements from bottom of the `segments` queue.
func (p *MediaPlaylist) mkM3u8() *bytes.Buffer {
	if p.buf.Len() > 0 {
		return &p.buf
	}

	p.addVersion()
	p.addCacheInfo()
	p.addPlaylistType()
	p.addTargetDuraion()
	p.addMediaSequence(uint64(1))
	p.addIframe()
	p.addSegments()
	p.buf.WriteString("#EXT-X-ENDLIST\n")
	return &p.buf
}

// Encode Generate output in live M3U8 format. Marshal `winsize` elements from bottom of the `segments` queue.
func (p *MediaPlaylist) mkLiveM3u8(sequence uint64) *bytes.Buffer {
	if p.buf.Len() > 0 {
		return &p.buf
	}

	p.addVersion()
	p.addCacheInfo()
	p.addTargetDuraion()
	p.addMediaSequence(sequence)
	p.addIframe()
	p.addSegments()
	return &p.buf
}

// String For compatibility with Stringer interface
// For example fmt.Printf("%s", sampleMediaList) will encode
// playist and print its string representation.
func (p *MediaPlaylist) String() string {
	return p.mkM3u8().String()
}

func (p *MediaPlaylist) LiveString(sequence uint64) string {
	return p.mkLiveM3u8(sequence).String()
}

func Mkm3u8(_segList []map[string]interface{}, _xl *xlog.Logger) string {
	length := len(_segList)
	pPlaylist := new(MediaPlaylist)
	pPlaylist.lastEndTime = -1
	pPlaylist.Init(uint(length), uint(length))
	_xl.Infof("length = %v", length)
	for _, v := range _segList {
		url := v["url"].(string)
		duration := v["duration"].(float64)
		pPlaylist.AppendSegment(url, duration, "", true)
	}
	return pPlaylist.String()
}

func MkLivem3u8(_segList []map[string]interface{}, sequence uint64, _xl *xlog.Logger) string {
	length := len(_segList)
	pPlaylist := new(MediaPlaylist)
	pPlaylist.lastEndTime = -1
	pPlaylist.Init(uint(length), uint(length))
	_xl.Infof("length = %v", length)
	for _, v := range _segList {
		url := v["url"].(string)
		duration := v["duration"].(float64)
		pPlaylist.AppendSegment(url, duration, "", false)
	}
	return pPlaylist.LiveString(sequence)
}
