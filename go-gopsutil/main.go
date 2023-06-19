package main

import (
	"log"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
)

func main() {
	// v, _ := mem.VirtualMemory()
	//
	// // almost every return value is a struct
	// fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)
	//
	// // convert to JSON. String() is also implemented
	// fmt.Println(v)

	println(cpu.Counts(true))
	infos, err := cpu.Info()
	if err != nil {
		log.Println(err)
	}

	for _, info := range infos {
		println(info.String())
	}

	processes, err := process.Processes()
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, p := range processes {
		println(p.String())
	}
}
