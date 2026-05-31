//go:build darwin

package pidpeek

import (
	"fmt"

	"github.com/ebitengine/purego"

	"github.com/xDarkicex/pidpeek/internal/darwin"
)

var _ Inspector = (*defaultInspector)(nil)

func Get(pid int) (Metrics, error) {
	if pid == 0 {
		return Metrics{}, ErrProcessNotFound
	}
	if err := darwin.EnsureInit(); err != nil {
		return Metrics{}, wrapErr("Get", pid, err)
	}
	m, err := darwin.ProcessMetrics(pid)
	if err != nil {
		return Metrics{}, wrapErr("Get", pid, err)
	}
	return Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum}, nil
}

func GetIdentity(pid int) (Identity, error) {
	if pid == 0 {
		return Identity{}, ErrProcessNotFound
	}
	if err := darwin.EnsureInit(); err != nil {
		return Identity{}, wrapErr("GetIdentity", pid, err)
	}
	id, err := darwin.ProcessIdentity(pid)
	if err != nil {
		return Identity{}, wrapErr("GetIdentity", pid, err)
	}
	return Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath}, nil
}

func GetAll(pid int) (ProcessInfo, error) {
	if pid == 0 {
		return ProcessInfo{}, ErrProcessNotFound
	}
	if err := darwin.EnsureInit(); err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, err)
	}
	m, err := darwin.ProcessMetrics(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, err)
	}
	id, err := darwin.ProcessIdentity(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, err)
	}
	return ProcessInfo{
		Metrics:  Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum},
		Identity: Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath},
	}, nil
}

func Self() (Metrics, error) {
	if err := darwin.EnsureInit(); err != nil {
		return Metrics{}, wrapErr("Self", 0, err)
	}
	m, err := darwin.ProcessMetrics(-1)
	if err != nil {
		return Metrics{}, wrapErr("Self", 0, err)
	}
	return Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum}, nil
}

func SelfIdentity() (Identity, error) {
	if err := darwin.EnsureInit(); err != nil {
		return Identity{}, wrapErr("SelfIdentity", 0, err)
	}
	id, err := darwin.ProcessIdentity(-1)
	if err != nil {
		return Identity{}, wrapErr("SelfIdentity", 0, err)
	}
	return Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath}, nil
}

func loadLibraries() error {
	libSystem, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("pidpeek: dlopen libSystem.B.dylib: %w", err)
	}
	libProc, err := purego.Dlopen("/usr/lib/libproc.dylib", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("pidpeek: dlopen libproc.dylib: %w", err)
	}
	purego.RegisterLibFunc(&darwin.MachTimebaseInfo, libSystem, "mach_timebase_info")
	purego.RegisterLibFunc(&darwin.ProcPidinfo, libProc, "proc_pidinfo")
	purego.RegisterLibFunc(&darwin.ProcPidpath, libProc, "proc_pidpath")

	var info darwin.MachTimebaseInfoData
	darwin.MachTimebaseInfo(&info)
	darwin.TimebaseRatioVal = float64(info.Numer) / float64(info.Denom)
	return nil
}

func init() {
	darwin.SetInitFunction(loadLibraries)
}