// Copyright 2013 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	flags struct {
		format     string
		separator  string
		widthEqual bool
	}
	cmd = "seq [-f format] [-w] [-s separator] [start] [step] <end>"
)

func usage() {
	fmt.Fprintf(os.Stdout, "Usage: %v\n", cmd)
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.Usage = usage
	flag.StringVar(&flags.format, "f", "%v", "use printf style floating-point FORMAT")
	flag.StringVar(&flags.separator, "s", "\n", "use STRING to separate numbers")
	flag.BoolVar(&flags.widthEqual, "w", false, "equalize width by padding with leading zeroes")
	flag.Parse()
}

func seq(w io.Writer, args []string) error {
	var (
		stt   float64 = 1.0
		stp   float64 = 1.0
		end   float64
		width int
	)

	format := flags.format // I use that because I'll modify a global variable
	argv, argc := args, len(args)
	if argc < 1 || argc > 4 {
		return errors.New(fmt.Sprintf("Mismatch n args; got %v, wants 1 >= n args >= 3", argc))
	}

	// loading step value if args is <start> <step> <end>
	if argc == 3 {
		_, err := fmt.Sscanf(argv[1], "%v", &stp)
		if stp-float64(int(stp)) > 0 && format == "%v" {
			d := len(fmt.Sprintf("%v", stp-float64(int(stp)))) - 2 // get the nums of y.xx decimal part
			format = fmt.Sprintf("%%.%df", d)
		}
		if stp == 0.0 {
			return errors.New("Step value should be != 0.")
		}

		if err != nil {
			return err
		}
	}

	if argc >= 2 { // cases: start + end || start + step + end
		if _, err := fmt.Sscanf(argv[0]+" "+argv[argc-1], "%v %v", &stt, &end); err != nil {
			return err
		}
	} else { // only <end>
		if _, err := fmt.Sscanf(argv[0], "%v", &end); err != nil {
			return err
		}
	}

	format = strings.Replace(format, "%", "%0*", 1) // support widthEqual
	if flags.widthEqual {
		width = len(fmt.Sprintf(format, 0, end))
	}

	defer fmt.Fprint(w, "\n") // last char is always '\n'
	for stt <= end {
		fmt.Fprintf(w, format, width, stt)
		stt += stp
		if stt <= end { // print only between the values
			fmt.Fprint(w, flags.separator)
		}
	}

	return nil
}

func main() {
	if err := seq(os.Stdout, flag.Args()); err != nil {
		log.Println(err)
		flag.Usage()
	}
}
