// Last Update:2018-07-05 10:07:20
/**
 * @file h264_parse.c
 * @brief pase the raw h264 stream 
 * @author liyq
 * @version 0.1.00
 * @date 2018-07-03
 */

#include <string.h>
#include "common.h"

/*
 * the NAL start prefix code (it can also be 0x00000001, depends on the encoder implementation)
 * */
#define NALU_START_CODE_4BYTES (0x01000000)
#define NALU_START_CODE_3BYTES (0x010000)

#define MAX_NALU_NUM_PER_FRAME 8

#define NALU_TYPE_NONIDR	1 // Coded slice of a non-IDR picture
// Coded slice data partition A
#define NALU_TYPE_SLICE_DPA 2 // P frame
#define NALU_TYPE_SLICE_DPB 3
#define NALU_TYPE_SLICE_DPC 4
#define NALU_TYPE_IDR		5 // I frame
// Supplemental enhancement information
#define NALU_TYPE_SEI		6
#define NALU_TYPE_SPS		7 // Sequence parameter set
#define NALU_TYPE_PPS		8 // Picture parameter set
// Access unit delimiter
#define NALU_TYPE_AUD       9
/*
 *  access unit delimiter. Notice that it is immediately followed by another NAL unit defined by 0x67,
 *  which is a NAL type of 7, which is the sequence parameter set
 *
 *  Access Unit Delimiter (AUD). An AUD is an optional NALU that can be use to delimit frames in an elementary stream.
 *  It is not required (unless otherwise stated by the container/protocol, like TS),
 *  and is often not included in order to save space, but it can be useful to finds the start of a frame without having to fully parse each NALU.
 *
 * */
#define NALU_TYPE_AUD		9 // 

typedef struct
{
	unsigned int timeStamp;
	unsigned short don;
	unsigned int size;
	unsigned char * addr;
	unsigned char naluType;
    unsigned char forbidden;
    unsigned char refRdc;
} NALU;

typedef struct {
    NALU nalus[MAX_NALU_NUM_PER_FRAME];
    int nonNALUCount;
    int NALUCount;
} H264Info;

static H264Info gH264Info;

// RBSP - Raw Byte Sequence Payload 
int RBSPParse( unsigned char *_pBitStream, int index )
{
    unsigned char *pBitStream = _pBitStream;

    gH264Info.nalus[index].naluType = (*pBitStream) & 0x1F;// 5 bit
    gH264Info.nalus[index].forbidden = (*pBitStream) & 0x80; // 1 bit
    gH264Info.nalus[index].refRdc = (*pBitStream) & 0x60;// 2 bit

    return 0;
}

int H264Parse2( unsigned char *_pBitStream, unsigned int size )
{
    unsigned char *pBitStream = _pBitStream;
    unsigned int *pStartCode = NULL;
    int index = 0;

    while( pBitStream <= _pBitStream ) {
        pStartCode = (unsigned int *)pBitStream;
        if ( *pStartCode == NALU_START_CODE_4BYTES ) {
            gH264Info.NALUCount++;
            pBitStream += 4;// skip start code
            RBSPParse( pBitStream, index );
        } else if ( *pStartCode == NALU_START_CODE_3BYTES ) {
            gH264Info.NALUCount++;
            pBitStream += 3;
            RBSPParse( pBitStream, index );
        } else if ( pBitStream[3] != 0x00 ) { 
            /* 
             *  ----------------------------------------------------------------
             * | x1 | x2 | x3 | x4 ( not 0x00 & not 0x01 ) | x5 | x6 | x7 | ... |
             *  ----------------------------------------------------------------
             * 
             * if the 4th byte is not 0x00 and not 0x01, so
             * 1. x2 x3 x4 x5
             * 2. x3 x4 x5 x6
             * 3. x4 x5 x6 x7
             * this 3 case can not be 00 00 00 01 or 00 00 01, it expect the x4 must been 0x00
             * so we can skip them
             */
            pBitStream += 4;
            gH264Info.nonNALUCount ++;
        } else if ( pBitStream[2] != 0x00 ) {
            pBitStream += 3;
            gH264Info.nonNALUCount ++;
        } else if ( pBitStream[1] != 0x00 ) {
            pBitStream += 2;
            gH264Info.nonNALUCount ++;
        } else {
            pBitStream += 1;
            gH264Info.nonNALUCount ++;
        }
        index ++;
    }

    return 0;
}


