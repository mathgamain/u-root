// package uroot contains various functions that might be needed more than
// one place.
package uroot

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	PATH = "/bin:/ubin:/buildbin:/usr/local/bin"
	CmdsPath = "github.com/u-root/u-root/cmds"
)

// TODO: make this a map so it's easier to find dups.
type dir struct {
	name string
	mode os.FileMode
}

type file struct {
	contents string
	mode     os.FileMode
}

// TODO: make this a map so it's easier to find dups.
type dev struct {
	name    string
	mode    os.FileMode
	magic   int
	howmany int
}
type mount struct {
	source string
	target string
	fstype string
	flags  uintptr
	opts   string
}

var (
	Envs []string
	env  = map[string]string{
		"LD_LIBRARY_PATH": "/usr/local/lib",
		"GOROOT":          "/go",
		"GOPATH":          "/",
		"CGO_ENABLED":     "0",
	}

	dirs = []dir{
		{name: "/proc", mode: os.FileMode(0555)},
		{name: "/sys", mode: os.FileMode(0555)},
		{name: "/buildbin", mode: os.FileMode(0777)},
		{name: "/ubin", mode: os.FileMode(0777)},
		{name: "/tmp", mode: os.FileMode(0777)},
		{name: "/env", mode: os.FileMode(0777)},
		{name: "/etc", mode: os.FileMode(0777)},
		{name: "/tcz", mode: os.FileMode(0777)},
		{name: "/dev", mode: os.FileMode(0777)},
		{name: "/lib", mode: os.FileMode(0777)},
		{name: "/usr/lib", mode: os.FileMode(0777)},
		{name: "/go/pkg/linux_amd64", mode: os.FileMode(0777)},
	}
	devs = []dev{
	// chicken and egg: these need to be there before you start. So, sadly,
	// we will always need dev.cpio.
	//{name: "/dev/null", mode: os.FileMode(0660) | 020000, magic: 0x0103},
	//{name: "/dev/console", mode: os.FileMode(0660) | 020000, magic: 0x0501},
	}
	namespace = []mount{
		{source: "proc", target: "/proc", fstype: "proc", flags: syscall.MS_MGC_VAL | syscall.MS_RDONLY, opts: ""},
		{source: "sys", target: "/sys", fstype: "sysfs", flags: syscall.MS_MGC_VAL | syscall.MS_RDONLY, opts: ""},
	}

	files = map[string]file{
		"/etc/resolv.conf": {contents: `nameserver 8.8.8.8`, mode: os.FileMode(0644)},
	}
)

// build the root file system.
func Rootfs() {
	// Pick some reasonable values in the (unlikely!) even that Uname fails.
	uname := "linux"
	mach := "x86_64"
	// There are three possible places for go:
	// The first is in /go/bin/$OS_$ARCH
	// The second is in /go/bin [why they still use this path is anyone's guess]
	// The third is in /go/pkg/tool/$OS_$ARCH
	if u, err := Uname(); err != nil {
		log.Printf("uroot.Utsname fails: %v, so assume %v_%v\n", err, uname, mach)
	} else {
		// Sadly, go and the OS disagree on case.
		uname = strings.ToLower(u.Sysname)
		mach = strings.ToLower(u.Machine)
		// Yes, we really have to do this stupid thing.
		if mach[0:3] == "arm" {
			mach = "arm"
		}
	}
	env["PATH"] = fmt.Sprintf("/go/bin/%s_%s:/go/bin:/go/pkg/tool/%s_%s:%v", uname, mach, uname, mach, PATH)

	for k, v := range env {
		os.Setenv(k, v)
		Envs = append(Envs, k+"="+v)
	}

	for _, m := range dirs {
		if err := os.MkdirAll(m.name, m.mode); err != nil {
			log.Printf("mkdir :%s: mode %o: %v\n", m.name, m.mode, err)
			continue
		}
	}

	for _, d := range devs {
		syscall.Unlink(d.name)
		if err := syscall.Mknod(d.name, uint32(d.mode), d.magic); err != nil {
			log.Printf("mknod :%s: mode %o: magic: %v: %v\n", d.name, d.mode, d.magic, err)
			continue
		}
	}

	for _, m := range namespace {
		if err := syscall.Mount(m.source, m.target, m.fstype, m.flags, m.opts); err != nil {
			log.Printf("Mount :%s: on :%s: type :%s: flags %x opts: %s: %v\n", m.source, m.target, m.fstype, m.flags, m.opts, err)
		}

	}

	for name, m := range files {
		if err := ioutil.WriteFile(name, []byte(m.contents), m.mode); err != nil {
			log.Printf("Error writeing %v: %v", name, err)
		}
	}

	// only in case of emergency.
	if false {
		if err := filepath.Walk("/", func(name string, fi os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf(" WALK FAIL%v: %v\n", name, err)
				// That's ok, sometimes things are not there.
				return nil
			}
			fmt.Printf("%v\n", name)
			return nil
		}); err != nil {
			log.Printf("WALK fails %v\n", err)
		}
	}
}
