package main

import "flag"
import "fmt"

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func scan(pid int, kw string) {
	proc := new(Proc)
	proc.Init(pid)

	fmt.Println("scanning heap ...")
	for addr := proc.HeapTop; addr < proc.HeapBottom; addr++ {
		mem, err := proc.ReadMem(addr, len(kw))
		if err != nil {
			continue
		}
		if string(mem) == kw {
			fmt.Printf("FOUND!!! addr: 0x%x\n", addr)
		}
	}
	fmt.Println("not found.")
}

func dump(pid int, addr int, sz int, hex bool) {
	proc := new(Proc)
	proc.Init(pid)

	mem, err := proc.ReadMem(uint64(addr), sz)
	if err != nil {
		fmt.Println("error reading mem")
		return
	}

	if hex {
		fmt.Println(mem)
	} else {
		fmt.Println(string(mem))
	}
	
}


func main() {
	pid := flag.Int("p", 0, "pid")
	kw := flag.String("k","", "keyword")
	a := flag.Int("a", 0, "address")
	bs := flag.Int("bs", 0, "number of bytes")
	x := flag.Bool("x", false, "display hex bytes")
	flag.Parse()

	if *pid > 0 && *kw != "" {
		scan(*pid, *kw)
	} else {
		if *pid > 0 && *a != 0 && *bs > 0 {
			dump(*pid, *a, *bs, *x)
		} else {
			fmt.Println("try --help")
		}
	}
}
