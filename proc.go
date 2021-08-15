package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// Proc process manipulation object
type Proc struct {
	Pid        int
	HeapTop    uint64
	HeapBottom uint64
}

// Init initialize proc object providing the pid
func (p *Proc) Init(pid int) {
	p.Pid = pid
	p.HeapTop = 0
	p.HeapBottom = 0
	p.getHeapData()
}

// Str object string
func (p Proc) Str() string {
	s := "pid: " + strconv.Itoa(p.Pid) + " heap top: " + strconv.FormatUint(p.HeapTop, 16) + " bottom: " + strconv.FormatUint(p.HeapBottom, 16)
	return s
}

// IsHeap verify that an address is in heap
func (p *Proc) IsHeap(addr uint64) bool {
	if p.HeapTop <= addr && addr <= p.HeapBottom {
		return true
	}
	return false
}

func (p *Proc) getHeapData() {
	maps := fmt.Sprintf("/proc/%d/maps", p.Pid)

	file, err := os.Open(maps)
	if err != nil {
		log.Println("wrong pid")
		file.Close()
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[heap]") {
			fmt.Sscanf(line, "%x-%x", &p.HeapTop, &p.HeapBottom)
			file.Close()
			return
		}
	}
	file.Close()
}

// ReadMem read mapped memory of current process
func (p Proc) ReadMem(addr uint64, sz int) ([]byte, error) {

	fnamemem := fmt.Sprintf("/proc/%d/mem", p.Pid)
	file, err := os.Open(fnamemem)
	if err != nil {
		file.Close()
		return nil, errors.New("wrong pid")
	}

	data := make([]byte, sz)
	n, err := file.ReadAt(data, int64(addr))
	if err != nil {
		file.Close()
		return nil, errors.New("Cannot read maps")
	}
	if n != sz {
		file.Close()
		return nil, errors.New("Incomplete read")
	}

	file.Close()

	return data, nil
}

// ReadMemPtr64 read memory getting an int64 of current process
func (p Proc) ReadMemPtr64(addr uint64) (uint64, error) {
	mem, err := p.ReadMem(addr, 8)
	if err != nil {
		return 0, err
	}
	ptr := binary.LittleEndian.Uint64(mem)
	return ptr, nil
}

// ReadMemUint32 read memory getting an int32 of current process
func (p Proc) ReadMemUint32(addr uint64) (uint32, error) {
	mem, err := p.ReadMem(addr, 8)
	if err != nil {
		return 0, err
	}
	n := binary.LittleEndian.Uint32(mem)
	return n, nil
}

// ReadMemInt32 read memory getting an int32 of current process
func (p Proc) ReadMemInt32(addr uint64) (int32, error) {
	mem, err := p.ReadMem(addr, 8)
	if err != nil {
		return 0, err
	}
	n := binary.LittleEndian.Uint32(mem)
	return int32(n), nil
}

// GetSocketInodes get all the inodes of current process
func (p Proc) GetSocketInodes() []uint32 {
	fdfolder := fmt.Sprintf("/proc/%d/fd", p.Pid)
	files, err := ioutil.ReadDir(fdfolder)
	if err != nil {
		log.Println("cant walk fds")
		return nil
	}

	var inode uint32
	var inodes []uint32

	for _, file := range files {

		linkname := fmt.Sprintf("%s/%s", fdfolder, file.Name())
		link, err := os.Readlink(linkname)
		if err != nil {
			log.Println(err)
			log.Println("cant read link " + file.Name())
			continue
		}

		if strings.Contains(link, "socket:") {
			fmt.Sscanf(link, "socket:[%d]", &inode)
			inodes = append(inodes, inode)
		}
	}

	return inodes
}

func Inode2Pid(inode uint32) (int, error) {
	processes, err := ioutil.ReadDir("/proc/")
	if err != nil {
		return 0, errors.New("this shouldnt happen")
	}

	for _, process := range processes {
		if !process.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(process.Name())
		if err != nil {
			continue // not a number
		}

		proc := new(Proc)
		proc.Init(pid)

		for _, pinode := range proc.GetSocketInodes() {
			if pinode == inode {
				return pid, nil
			}
		}
	}

	fmt.Println("not found")
	return 0, errors.New("process not found")
}

// Inodes2Pid having a list of inodes find the pid/pids with this inodes (normally one pid)
func Inodes2Pid(inodes []uint32) (int, error) {

	processes, err := ioutil.ReadDir("/proc/")
	if err != nil {
		return 0, errors.New("this shouldnt happen")
	}

	for _, process := range processes {
		if !process.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(process.Name())
		if err != nil {
			continue // not a number
		}

		proc := new(Proc)
		proc.Init(pid)
		for _, pinode := range proc.GetSocketInodes() {
			for _, inode := range inodes {
				if pinode == inode {
					return pid, nil
				}
			}
		}
	}

	return 0, errors.New("process not found")
}

// Pid2Name retrieve the processname of a pid
func Pid2Name(pid int) string {
	comm := fmt.Sprintf("/proc/%d/comm", pid)
	data, err := ioutil.ReadFile(comm)
	if err != nil {
		return "??"
	}
	return string(data)
}

// ProcLocatePids locate the pids with a specific processname
func ProcLocatePids(procname string) []int {
	var pids []int

	files, err := ioutil.ReadDir("/proc/")
	Check(err)
	for _, pid := range files {

		cmdlinefile := fmt.Sprintf("/proc/%s/cmdline", pid.Name())
		cmdline, err := ioutil.ReadFile(cmdlinefile)
		if err != nil {
			continue
		}
		if strings.Contains(string(cmdline), procname) {
			ipid, err := strconv.Atoi(pid.Name())
			if err != nil {
				continue
			}
			pids = append(pids, ipid)
		}
	}

	return pids
}
