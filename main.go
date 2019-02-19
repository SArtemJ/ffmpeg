package main

import (
	"fmt"
	"github.com/asticode/goav/avcodec"
	"github.com/asticode/goav/avformat"
	"github.com/asticode/goav/avutil"
)

func main() {

	// Alloc ctx
	ctxFormat := avformat.AvformatAllocContext()
	var ret int

	// Open input
	// We need to create an intermediate variable to avoid "cgo argument has Go pointer to Go pointer" errors
	if ret = avformat.AvformatOpenInput(&ctxFormat, "sample.mp4", nil, nil); ret < 0 {
		fmt.Printf("astilibav: avformat.AvformatOpenInput on %+v failed", ret)
		return
	}

	//info about streams in file
	if ret = ctxFormat.AvformatFindStreamInfo(nil); ret < 0 {
		fmt.Printf("astilibav: avformat.AvformatOpenInput on %+v failed", ret)
		return
	}

	for _, stream := range ctxFormat.Streams() {
		var codecParams *avcodec.CodecParameters
		codecParams = stream.CodecParameters()

		var ctxCodec *avcodec.Context

		// Find decoder
		var cdc *avcodec.Codec
		if cdc = avcodec.AvcodecFindDecoder(codecParams.CodecId()); cdc == nil {
			fmt.Printf("no decoder found for codec")
			return
		}

		// Alloc context
		if ctxCodec = cdc.AvcodecAllocContext3(); ctxCodec == nil {
			fmt.Printf("no context allocated for codec %+v", cdc)
			return
		}

		// Copy codec parameters
		if ret := avcodec.AvcodecParametersToContext(ctxCodec, codecParams); ret < 0 {
			fmt.Printf("avcodec.AvcodecParametersToContext failed")
			return
		}

		// Open codec
		if ret := ctxCodec.AvcodecOpen2(cdc, nil); ret < 0 {
			fmt.Printf("ctxCodec.AvcodecOpen2 failed")
			return
		}

		var pkt *avcodec.Packet
		if ret := avcodec.AvcodecSendPacket(ctxCodec, pkt); ret < 0 {
			fmt.Printf("avcodec.AvcodecSendPacket failed")
			return
		}

		if ret > 0 {
			f := avutil.AvFrameAlloc()
			if ret := avcodec.AvcodecReceiveFrame(ctxCodec, f); ret < 0 {
				if ret != avutil.AVERROR_EOF && ret != avutil.AVERROR_EAGAIN {
					fmt.Printf("avcodec.AvcodecReceiveFrame failed")
					return
				}
				return
			}

			ctxOut := avformat.AvformatAllocContext()
			if ret = avformat.AvformatOpenInput(&ctxFormat, "sample.ts", nil, nil); ret < 0 {
				fmt.Printf("astilibav: avformat.AvformatOpenInput on %+v failed", ret)
				return
			}

			// Find encoder
			var cdc *avcodec.Codec
			if cdc = avcodec.AvcodecFindEncoderByName("ts"); cdc == nil {
				fmt.Printf(" no encoder with name")
				return
			}

			// Alloc context
			if ctxCodec := cdc.AvcodecAllocContext3(); ctxCodec == nil {
				fmt.Printf("  no context allocated")
				return
			}

			// Set shared context parameters
			if ctxOut.GlobalHeader {
				ctxCodec.SetFlags(ctxCodec.Flags() | avcodec.AV_CODEC_FLAG_GLOBAL_HEADER)
			}
			if ctxOut.ThreadCount != nil {
				ctxCodec.SetThreadCount(*ctxOut.ThreadCount)
			}

			// Set media type-specific context parameters
			switch ctxOut.CodecType {
			case avutil.AVMEDIA_TYPE_AUDIO:
				ctxCodec.SetBitRate(int64(ctxOut.BitRate))
				ctxCodec.SetChannelLayout(ctxOut.ChannelLayout)
				ctxCodec.SetChannels(ctxOut.Channels)
				ctxCodec.SetSampleFmt(ctxOut.SampleFmt)
				ctxCodec.SetSampleRate(ctxOut.SampleRate)
			case avutil.AVMEDIA_TYPE_VIDEO:
				ctxCodec.SetBitRate(int64(ctxOut.BitRate))
				ctxCodec.SetFramerate(ctxOut.FrameRate)
				ctxCodec.SetGopSize(ctxOut.GopSize)
				ctxCodec.SetHeight(ctxOut.Height)
				ctxCodec.SetPixFmt(ctxOut.PixelFormat)
				ctxCodec.SetSampleAspectRatio(ctxOut.SampleAspectRatio)
				ctxCodec.SetTimeBase(ctxOut.TimeBase)
				ctxCodec.SetWidth(ctxOut.Width)
			default:
				fmt.Printf("  encoder doesn't handle %v codec type", ctxOut.CodecType)
				return
			}

			// Open codec
			var dict *avutil.Dictionary
			if ret := ctxCodec.AvcodecOpen2(cdc, dict); ret < 0 {
				fmt.Printf("  ctxCodec.AvcodecOpen2 failed")
				return
			}

			frameResult := avutil.AvFrameAlloc()
			if frameResult != nil {
				switch ctxCodec.CodecType() {
				case avutil.AVMEDIA_TYPE_VIDEO:
					frameResult.SetKeyFrame(0)
					frameResult.SetPictType(avutil.AvPictureType(avutil.AV_PICTURE_TYPE_NONE))
				}
			}

			// Send frame to encoder
			if ret := avcodec.AvcodecSendFrame(ctxCodec, frameResult); ret < 0 {
				fmt.Printf("  avcodec.AvcodecSendFrame failed")
				return
			}

			if ret > 0 {
				var pktRes *avformat.Packet

				if ret := avcodec.AvcodecReceivePacket(ctxCodec, pktRes); ret < 0 {
					if ret != avutil.AVERROR_EOF && ret != avutil.AVERROR_EAGAIN {
						fmt.Printf(" avcodec.AvcodecReceivePacket failed")
					}
					return
				}

				// Set pkt duration based on framerate
				if f := ctxCodec.Framerate(); frameResult.Num() > 0 {
					pkt.SetDuration(avutil.AvRescaleQ(int64(1e9/f.ToDouble()), avutil.NewRational(1, 1e9), ctxFormat.TimeBase()))
				}

				// Rescale timestamps
				pkt.AvPacketRescaleTs(ctxFormat.TimeBase(), ctxCodec.TimeBase())

				return
			}

		}

	}
}
