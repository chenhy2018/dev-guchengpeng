// Last Update:2018-07-03 19:08:11
/**
 * @file h264_parse.c
 * @brief pase the raw h264 stream 
 * @author liyq
 * @version 0.1.00
 * @date 2018-07-03
 */

#include <string.h>
#include "common.h"

#define MAX_NALU_NUM_PER_FRAME 8

#define NALU_TYPE_NONIDR	1
#define NALU_TYPE_SLICE_DPA 2 // P frame
#define NALU_TYPE_SLICE_DPB 3
#define NALU_TYPE_SLICE_DPC 4
#define NALU_TYPE_IDR		5 // I frame
#define NALU_TYPE_SEI		6
#define NALU_TYPE_SPS		7 // Sequence parameter set
#define NALU_TYPE_PPS		8 // Picture parameter set
#define NALU_TYPE_AUD		9

typedef struct
{
	unsigned int timeStamp;
	unsigned short don;
	unsigned int size;
	unsigned char * addr;
	unsigned char nalu_type;
} NALU;


int H265GetNALUType(unsigned char *pBuffer)
{
    unsigned char data = pBuffer[0];
    int nanutype = (data >> 1) & 0x3f;

    return nanutype;
}

int H264GetNALUType(unsigned char *pBuffer)
{
    unsigned char data = pBuffer[0];
    int nanutype = (data) & 0x1F;

    return nanutype;
}

int H264Parse( int encodetype, unsigned char *bitstream, unsigned int streamSize )
{
    NALU nalus[MAX_NALU_NUM_PER_FRAME];
    int index = -1;
    u_int8_t * bs = bitstream;
    u_int32_t head;
    u_int8_t nalu_type;
    int count = 0;
    u_int8_t *last_byte_bitstream = bitstream + streamSize - 1;

    memset(nalus, 0, sizeof(nalus));

    while ( bs <= last_byte_bitstream ) {

        head = (bs[3] << 24) | (bs[2] << 16) | (bs[1] << 8) | bs[0];

        if (head == 0x01000000) {	// little ending
            index++;
            // we find a nalu
            bs += 4;		// jump to nalu type

            if( encodetype == 264 )
                nalu_type = H264GetNALUType(bs);
            else
                nalu_type = H265GetNALUType(bs);
            nalus[index].nalu_type = nalu_type;
            nalus[index].addr = bs;

            if (index  > 0) {	// Not the first NALU in this stream
                nalus[index -1].size = nalus[index].addr - nalus[index -1].addr - 4; // cut off 4 bytes of delimiter
            }
            //			printf("	nalu type %d, index %d, previous size %d\n", nalus[index].nalu_type, index,  index > 0 ? nalus[index -1].size : 0);
        }
        else if (bs[3] != 0) {
            bs += 4;
        } else if (bs[2] != 0) {
            bs += 3;
        } else if (bs[1] != 0) {
            bs += 2;
        } else {
            bs += 1;
        }
    }

    if( index >= 0)
        nalus[index].size =  last_byte_bitstream - nalus[index].addr + 1;

    count = index + 1;
    if (count == 0)
    {
        if (streamSize == 0)
        {
//            debuglog(SLOG_LVL_TRACE, "stream ended.\n");
        }
        else
        {
//            debuglog(SLOG_LVL_TRACE, "No nalu found in the bitstream!\n");
        }
        return -1;
    }


    int j = 0;
    for( j = 0; j< count; j++)
    {
        if( (NALU_TYPE_SPS == nalus[j].nalu_type && 264 == encodetype ) ||
            (NALU_TYPE_PPS == nalus[j].nalu_type && 264 == encodetype ) ||
            (32 == nalus[j].nalu_type && 265 == encodetype ) ||
            (33 == nalus[j].nalu_type && 265 == encodetype ) ||
            (34 == nalus[j].nalu_type && 265 == encodetype ) )
        {
//            debuglog(SLOG_LVL_TRACE, "index %d: type %d, offset %d, size %d.\n", j, nalus[j].nalu_type, nalus[j].addr - bitstream, nalus[j].size);
//            deubug_show_data_hex((unsigned char *) nalus[j].addr-4, nalus[j].size+4, 0);
        }


        if( (NALU_TYPE_SEI == nalus[j].nalu_type && 264 == encodetype ) ||
            (/*H265_NAL_UNIT_SEI*/39 == nalus[j].nalu_type && 265 == encodetype ) ||
            (/*H265_NAL_UNIT_SEI_SUFFIX*/40 == nalus[j].nalu_type && 265 == encodetype ) )
        {
//            debuglog(SLOG_LVL_TRACE, "index %d: type %d, offset %d, size %d.\n", j, nalus[j].nalu_type, nalus[j].addr - bitstream, nalus[j].size);
//            deubug_show_data_hex((unsigned char *) nalus[j].addr-4, nalus[j].size+4, 0);
        }
    }

    return 0;
}

