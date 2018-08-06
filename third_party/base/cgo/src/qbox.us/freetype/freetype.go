package freetype

/*
#cgo CFLAGS: -I/usr/local/include/freetype2 -I/usr/include/freetype2
#cgo LDFLAGS: -L/usr/local/lib -lfreetype

#include <stdlib.h>
#include <ft2build.h>
#include FT_FREETYPE_H
*/
import "C"
import (
	"strconv"
	"unsafe"
)

// --------------------------------------------------------------------

const (
	StyleFlag_Italic = 1
	StyleFlag_Bold   = 2
)

// --------------------------------------------------------------------
// type Errno

type Errno C.FT_Error

func (r Errno) Error() string {

	if desc, ok := g_errdesc[int(r)]; ok {
		return desc
	}
	return "E" + strconv.Itoa(int(r))
}

var g_errdesc = map[int]string{

	C.FT_Err_Ok:                    "no error",
	C.FT_Err_Cannot_Open_Resource:  "cannot open resource",
	C.FT_Err_Unknown_File_Format:   "unknown file format",
	C.FT_Err_Invalid_File_Format:   "broken file",
	C.FT_Err_Invalid_Version:       "invalid FreeType version",
	C.FT_Err_Lower_Module_Version:  "module version is too low",
	C.FT_Err_Invalid_Argument:      "invalid argument",
	C.FT_Err_Unimplemented_Feature: "unimplemented feature",
	C.FT_Err_Invalid_Table:         "broken table",
	C.FT_Err_Invalid_Offset:        "broken offset within table",
	C.FT_Err_Array_Too_Large:       "array allocation size too large",

	C.FT_Err_Invalid_Glyph_Index:    "invalid glyph index",
	C.FT_Err_Invalid_Character_Code: "invalid character code",
	C.FT_Err_Invalid_Glyph_Format:   "unsupported glyph image format",
	C.FT_Err_Cannot_Render_Glyph:    "cannot render this glyph format",
	C.FT_Err_Invalid_Outline:        "invalid outline",
	C.FT_Err_Invalid_Composite:      "invalid composite glyph",
	C.FT_Err_Too_Many_Hints:         "too many hints",
	C.FT_Err_Invalid_Pixel_Size:     "invalid pixel size",

	C.FT_Err_Invalid_Handle:         "invalid object handle",
	C.FT_Err_Invalid_Library_Handle: "invalid library handle",
	C.FT_Err_Invalid_Driver_Handle:  "invalid module handle",
	C.FT_Err_Invalid_Face_Handle:    "invalid face handle",
	C.FT_Err_Invalid_Size_Handle:    "invalid size handle",
	C.FT_Err_Invalid_Slot_Handle:    "invalid glyph slot handle",
	C.FT_Err_Invalid_CharMap_Handle: "invalid charmap handle",
	C.FT_Err_Invalid_Cache_Handle:   "invalid cache manager handle",
	C.FT_Err_Invalid_Stream_Handle:  "invalid stream handle",

	/*	C.FT_Err_Too_Many_Drivers,                            0x30, \
	                "too many modules" )
		C.FT_Err_Too_Many_Extensions,                         0x31, \
	                "too many extensions" )

		C.FT_Err_Out_Of_Memory,                               0x40, \
	                "out of memory" )
		C.FT_Err_Unlisted_Object,                             0x41, \
	                "unlisted object" )

		C.FT_Err_Cannot_Open_Stream,                          0x51, \
	                "cannot open stream" )
		C.FT_Err_Invalid_Stream_Seek,                         0x52, \
	                "invalid stream seek" )
		C.FT_Err_Invalid_Stream_Skip,                         0x53, \
	                "invalid stream skip" )
		C.FT_Err_Invalid_Stream_Read,                         0x54, \
	                "invalid stream read" )
		C.FT_Err_Invalid_Stream_Operation,                    0x55, \
	                "invalid stream operation" )
		C.FT_Err_Invalid_Frame_Operation,                     0x56, \
	                "invalid frame operation" )
		C.FT_Err_Nested_Frame_Access,                         0x57, \
	                "nested frame access" )
		C.FT_Err_Invalid_Frame_Read,                          0x58, \
	                "invalid frame read" )

		C.FT_Err_Raster_Uninitialized,                        0x60, \
	                "raster uninitialized" )
		C.FT_Err_Raster_Corrupted,                            0x61, \
	                "raster corrupted" )
		C.FT_Err_Raster_Overflow,                             0x62, \
	                "raster overflow" )
		C.FT_Err_Raster_Negative_Height,                      0x63, \
	                "negative height while rastering" )

		C.FT_Err_Too_Many_Caches,                             0x70, \
	                "too many registered caches" )

		C.FT_Err_Invalid_Opcode,                              0x80, \
	                "invalid opcode" )
		C.FT_Err_Too_Few_Arguments,                           0x81, \
	                "too few arguments" )
		C.FT_Err_Stack_Overflow,                              0x82, \
	                "stack overflow" )
		C.FT_Err_Code_Overflow,                               0x83, \
	                "code overflow" )
		C.FT_Err_Bad_Argument,                                0x84, \
	                "bad argument" )
		C.FT_Err_Divide_By_Zero,                              0x85, \
	                "division by zero" )
		C.FT_Err_Invalid_Reference,                           0x86, \
	                "invalid reference" )
		C.FT_Err_Debug_OpCode,                                0x87, \
	                "found debug opcode" )
		C.FT_Err_ENDF_In_Exec_Stream,                         0x88, \
	                "found ENDF opcode in execution stream" )
		C.FT_Err_Nested_DEFS,                                 0x89, \
	                "nested DEFS" )
		C.FT_Err_Invalid_CodeRange,                           0x8A, \
	                "invalid code range" )
		C.FT_Err_Execution_Too_Long,                          0x8B, \
	                "execution context too long" )
		C.FT_Err_Too_Many_Function_Defs,                      0x8C, \
	                "too many function definitions" )
		C.FT_Err_Too_Many_Instruction_Defs,                   0x8D, \
	                "too many instruction definitions" )
		C.FT_Err_Table_Missing,                               0x8E, \
	                "SFNT font table missing" )
		C.FT_Err_Horiz_Header_Missing,                        0x8F, \
	                "horizontal header (hhea) table missing" )
		C.FT_Err_Locations_Missing,                           0x90, \
	                "locations (loca) table missing" )
		C.FT_Err_Name_Table_Missing,                          0x91, \
	                "name table missing" )
		C.FT_Err_CMap_Table_Missing,                          0x92, \
	                "character map (cmap) table missing" )
		C.FT_Err_Hmtx_Table_Missing,                          0x93, \
	                "horizontal metrics (hmtx) table missing" )
		C.FT_Err_Post_Table_Missing,                          0x94, \
	                "PostScript (post) table missing" )
		C.FT_Err_Invalid_Horiz_Metrics,                       0x95, \
	                "invalid horizontal metrics" )
		C.FT_Err_Invalid_CharMap_Format,                      0x96, \
	                "invalid character map (cmap) format" )
		C.FT_Err_Invalid_PPem,                                0x97, \
	                "invalid ppem value" )
		C.FT_Err_Invalid_Vert_Metrics,                        0x98, \
	                "invalid vertical metrics" )
		C.FT_Err_Could_Not_Find_Context,                      0x99, \
	                "could not find context" )
		C.FT_Err_Invalid_Post_Table_Format,                   0x9A, \
	                "invalid PostScript (post) table format" )
		C.FT_Err_Invalid_Post_Table,                          0x9B, \
	                "invalid PostScript (post) table" )

		C.FT_Err_Syntax_Error,                                0xA0, \
	                "opcode syntax error" )
		C.FT_Err_Stack_Underflow,                             0xA1, \
	                "argument stack underflow" )
		C.FT_Err_Ignore,                                      0xA2, \
	                "ignore" )

		C.FT_Err_Missing_Startfont_Field,                     0xB0, \
	                "`STARTFONT' field missing" )
		C.FT_Err_Missing_Font_Field,                          0xB1, \
	                "`FONT' field missing" )
		C.FT_Err_Missing_Size_Field,                          0xB2, \
	                "`SIZE' field missing" )
		C.FT_Err_Missing_Chars_Field,                         0xB3, \
	                "`CHARS' field missing" )
		C.FT_Err_Missing_Startchar_Field,                     0xB4, \
	                "`STARTCHAR' field missing" )
		C.FT_Err_Missing_Encoding_Field,                      0xB5, \
	                "`ENCODING' field missing" )
		C.FT_Err_Missing_Bbx_Field,                           0xB6, \
	                "`BBX' field missing" )
		C.FT_Err_Bbx_Too_Big,                                 0xB7, \
	                "`BBX' too big" )
		C.FT_Err_Corrupted_Font_Header,                       0xB8, \
	                "Font header corrupted or missing fields" )
		C.FT_Err_Corrupted_Font_Glyphs,                       0xB9, \
	                "Font glyphs corrupted or missing fields" ) */
}

// --------------------------------------------------------------------
// type Library

type Library struct {
	Impl C.FT_Library
}

func New() (r Library, err error) {

	var lib C.FT_Library
	var ret = C.FT_Init_FreeType(&lib)
	if ret != 0 {
		err = Errno(ret)
		return
	}
	r = Library{lib}
	return
}

func (r Library) Release() (err error) {

	var ret = C.FT_Done_FreeType(r.Impl)
	if ret != 0 {
		err = Errno(ret)
	}
	return
}

// --------------------------------------------------------------------
// type Face

type Face struct {
	Impl C.FT_Face
}

func (r Library) NewFace(fontFile string, index int) (f Face, err error) {

	var face C.FT_Face
	var fname = C.CString(fontFile)
	var ret = C.FT_New_Face(r.Impl, fname, (C.FT_Long)(index), &face)
	C.free(unsafe.Pointer(fname))
	if ret != 0 {
		err = Errno(ret)
		return
	}
	f = Face{face}
	return
}

func (r Face) Release() (err error) {

	var ret = C.FT_Done_Face(r.Impl)
	if ret != 0 {
		err = Errno(ret)
	}
	return
}

type FaceRec struct {
	FamilyName string
	StyleName  string
	NumFaces   int
	FaceIndex  int
	FaceFlags  int
	StyleFlags int
	NumGlyphs  int
}

func (r Face) Detail() *FaceRec {

	var impl = r.Impl
	return &FaceRec{
		NumFaces:   int(impl.num_faces),
		FaceIndex:  int(impl.face_index),
		FaceFlags:  int(impl.face_flags),
		StyleFlags: int(impl.style_flags),
		NumGlyphs:  int(impl.num_glyphs),
		FamilyName: C.GoString((*C.char)(impl.family_name)),
		StyleName:  C.GoString((*C.char)(impl.style_name)),
	}
}

// --------------------------------------------------------------------
