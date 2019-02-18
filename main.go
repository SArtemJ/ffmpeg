package main

import (
	"fmt"
	"github.com/asticode/goav/avcodec"
	"github.com/asticode/goav/avformat"
	"github.com/asticode/goav/avutil"
)

func main() {

	// Alloc ctx
	ctxFormatIn := avformat.AvformatAllocContext()
	ctxFormatOut := avformat.AvformatAllocContext()

	var pkt *avformat.Packet
	var ret int

	// Open input
	// We need to create an intermediate variable to avoid "cgo argument has Go pointer to Go pointer" errors
	if ret := avformat.AvformatOpenInput(&ctxFormatIn, "sample.mp4", nil, nil); ret < 0 {
		fmt.Sprintf("astilibav: avformat.AvformatOpenInput on %+v failed", ret)
		return
	}

	//info about streams in file
	if ret := ctxFormatIn.AvformatFindStreamInfo(nil); ret < 0 {
		fmt.Sprintf("astilibav: avformat.AvformatOpenInput on %+v failed", ret)
		return
	}

	if ret := avformat.AvformatAllocOutputContext2(&ctxFormatOut, nil, "", "result.ts"); ret < 0 {
		fmt.Printf("avformat.AvformatAllocOutputContext2 on %+v failed")
		return
	}

	//print detailed info about file
	//ctxFormat.AvDumpFormat(0, "", 0)

	//range for all streams
	for _, stream := range ctxFormatIn.Streams() {
		var avStreamOut *avformat.Stream

		codecParamsType := stream.CodecParameters().CodecType()
		if codecParamsType != avcodec.AVMEDIA_TYPE_AUDIO && codecParamsType != avcodec.AVMEDIA_TYPE_VIDEO {
			continue
		}

		// Add stream
		avStreamOut = ctxFormatOut.AvformatNewStream(nil)

		// Set codec parameters
		if ret := avcodec.AvcodecParametersCopy(avStreamOut.CodecParameters(), stream.CodecParameters()); ret < 0 {
			fmt.Printf("avcodec.AvcodecParametersCopy from %+v to %+v failed")
			return
		}
	}

	if ctxFormatOut.Oformat().Flags()&avformat.AVFMT_NOFILE > 0 {
		var ctxAvIO *avformat.AvIOContext
		if ret := avformat.AvIOOpen(&ctxAvIO, "result.ts", avformat.AVIO_FLAG_WRITE); ret < 0 {
			fmt.Printf(" avformat.AvIOOpen on %+v failed")
			return
		}
	}

	var dict **avutil.Dictionary
	if ret := ctxFormatOut.AvformatWriteHeader(dict); ret < 0 {
		fmt.Printf("m.ctxFormat.AvformatWriteHeader on %s failed")
		return
	}

	for {
		var streamOut, streamIn *avformat.Stream
		if ret := ctxFormatIn.AvReadFrame(pkt); ret < 0 {
			break
		}

		streamIn = ctxFormatIn.Streams()[pkt.StreamIndex()]
		if pkt.StreamIndex() >= len(ctxFormatIn.Streams()) {
			pkt.AvPacketUnref()
			continue
		}

		streamOut = ctxFormatOut.Streams()[pkt.StreamIndex()]

		pts := avutil.AvRescaleQRnd(pkt.Pts(), streamIn.TimeBase(), streamOut.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX)
		pkt.SetPts(pts)

		dts := avutil.AvRescaleQRnd(pkt.Dts(), streamIn.TimeBase(), streamOut.TimeBase(), avutil.AV_ROUND_NEAR_INF|avutil.AV_ROUND_PASS_MINMAX)
		pkt.SetDts(dts)

		duration := avutil.AvRescaleQ(pkt.Duration(), streamIn.TimeBase(), streamOut.TimeBase())
		pkt.SetDuration(duration)

		pkt.SetPos(-1)

		if ret := ctxFormatOut.AvInterleavedWriteFrame(pkt); ret < 0 {
			fmt.Printf("m.ctxFormat.AvInterleavedWriteFrame on %s failed")
			return
		}

		if ret := ctxFormatOut.AvWriteTrailer(); ret < 0 {
			fmt.Printf("m.ctxFormat.AvWriteTrailer on %s failed")
			return
		}

	}

	avformat.AvformatCloseInput(ctxFormatIn)
	if ctxFormatOut.Oformat().Flags()&avformat.AVFMT_NOFILE > 0 {
		var av *avformat.AvIOContext
		av = ctxFormatOut.Pb()
		if ret := avformat.AvIOClosep(*av); ret < 0 {
			fmt.Printf("ctxFormat.AvIOClose on %s failed")
			return
		}

		ctxFormatOut.AvformatFreeContext()
	}

	if ret < 0 && ret != avutil.AVERROR_EOF {
		fmt.Printf("Error occurred: %v\n", ret)
		return
	}

	//codecType := stream.CodecParameters().CodecType()
	//fmt.Printf("codec params %+v\n", codecID)
	//fmt.Printf("codec params %+v\n", codecType)

	// Find decoder
	//var cdc *avcodec.Codec
	//if cdc = avcodec.AvcodecFindDecoder(stream.CodecParameters().CodecId()); cdc == nil {
	//	fmt.Printf("no decoder found for codec id %+v", stream.CodecParameters().CodecId())
	//		return
	//	}

	//fmt.Printf("codec info %+v", cdc)

	//var ctxCodec *avcodec.Context
	// Alloc context
	//if ctxCodec = cdc.AvcodecAllocContext3(); ctxCodec == nil {
	//	fmt.Printf("no context allocated for codec %+v", cdc)
	//	return
	//}

	// Copy codec parameters
	//if ret := avcodec.AvcodecParametersToContext(ctxCodec, stream.CodecParameters()); ret < 0 {
	//	fmt.Printf(" avcodec.AvcodecParametersToContext failed")
	//	return
	//}

	// Open codec
	//if ret := ctxCodec.AvcodecOpen2(cdc, nil); ret < 0 {
	//	fmt.Printf( "ctxCodec.AvcodecOpen2 failed")
	//	return

	// Make sure the codec is closed
	//c.Add(func() error {
	//	if ret := d.ctxCodec.AvcodecClose(); ret < 0 {
	//		emitAvError(nil, eh, ret, "d.ctxCodec.AvcodecClose failed")
	//	}
	//	return nil
	//})

	//	var pkt *avcodec.Packet
	//pkt = avcodec.AvPacketAlloc()

	//	var f *avutil.Frame
	//f = avutil.AvFrameAlloc()

	//if ret := ctxFormat.AvReadFrame(pkt); ret < 0 {
	//d.statWorkRatio.Done(true)
	//if ret != avutil.AVERROR_EOF || !d.loop {
	//	if ret != avutil.AVERROR_EOF {
	//		emitAvError(d, d.eh, ret, "ctxFormat.AvReadFrame on %s failed", d.ctxFormat.Filename())
	//	}
	//	stop = true
	//} else if d.loopFirstPkt != nil {
	//	// Seek to first pkt
	//	if ret = d.ctxFormat.AvSeekFrame(d.loopFirstPkt.s.Index(), d.loopFirstPkt.dts, avformat.AVSEEK_FLAG_BACKWARD); ret < 0 {
	//		emitAvError(d, d.eh, ret, "ctxFormat.AvSeekFrame on %s with stream idx %v and ts %v failed", d.ctxFormat.Filename(), d.loopFirstPkt.s.Index(), d.loopFirstPkt.dts)
	//		stop = true
	//	}
	//}
	//return
	//}
}
