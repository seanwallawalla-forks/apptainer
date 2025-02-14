// Copyright (c) 2021 Apptainer a Series of LF Projects LLC
//   For website terms of use, trademark policy, privacy policy and other
//   project policies see https://lfprojects.org/policies
// Copyright (c) 2018-2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Definition describes how to build an image.
type Definition struct {
	Header     map[string]string `json:"header"`
	ImageData  `json:"imageData"`
	BuildData  Data              `json:"buildData"`
	CustomData map[string]string `json:"customData"`
	Raw        []byte            `json:"raw"`
	// SCIF app sections must be processed in order from the definition file,
	// so we need to record the order of the items as they are parsed from the
	// file into unordered maps.
	AppOrder []string `json:"appOrder"`
}

// ImageData contains any scripts, metadata, etc... that needs to be
// present in some form in the final built image.
type ImageData struct {
	Metadata     []byte            `json:"metadata"`
	Labels       map[string]string `json:"labels"`
	ImageScripts `json:"imageScripts"`
}

// ImageScripts contains scripts that are used after build time.
type ImageScripts struct {
	Help        Script `json:"help"`
	Environment Script `json:"environment"`
	Runscript   Script `json:"runScript"`
	Test        Script `json:"test"`
	Startscript Script `json:"startScript"`
}

// Data contains any scripts, metadata, etc... that the Builder may
// need to know only at build time to build the image.
type Data struct {
	Files   []Files `json:"files"`
	Scripts `json:"buildScripts"`
}

// Scripts defines scripts that are used at build time.
type Scripts struct {
	Pre   Script `json:"pre"`
	Setup Script `json:"setup"`
	Post  Script `json:"post"`
	Test  Script `json:"test"`
}

// Files describes a %files section of a definition.
type Files struct {
	Args  string          `json:"args"`
	Files []FileTransport `json:"files"`
}

// FileTransport holds source and destination information of files to copy into the container.
type FileTransport struct {
	Src string `json:"source"`
	Dst string `json:"destination"`
}

// Script describes any script section of a definition.
type Script struct {
	Args   string `json:"args"`
	Script string `json:"script"`
}

// NewDefinitionFromURI crafts a new Definition given a URI.
func NewDefinitionFromURI(uri string) (d Definition, err error) {
	var u []string
	if strings.Contains(uri, "://") {
		u = strings.SplitN(uri, "://", 2)
	} else if strings.Contains(uri, ":") {
		u = strings.SplitN(uri, ":", 2)
	} else {
		return d, fmt.Errorf("build URI must start with prefix:// or prefix: ")
	}

	d = Definition{
		Header: map[string]string{
			"bootstrap": u[0],
			"from":      u[1],
		},
	}

	var buf bytes.Buffer
	populateRaw(&d, &buf)
	d.Raw = buf.Bytes()

	return d, nil
}

// NewDefinitionFromJSON creates a new Definition using the supplied JSON.
func NewDefinitionFromJSON(r io.Reader) (d Definition, err error) {
	decoder := json.NewDecoder(r)

	for {
		if err = decoder.Decode(&d); err == io.EOF {
			break
		} else if err != nil {
			return
		}
	}

	// if JSON definition doesn't have a raw data section, add it
	if len(d.Raw) == 0 {
		var buf bytes.Buffer
		populateRaw(&d, &buf)
		d.Raw = buf.Bytes()
	}

	return d, nil
}

func writeSectionIfExists(w io.Writer, ident string, s Script) {
	if len(s.Script) > 0 {
		fmt.Fprintf(w, "%%%s", ident)
		if len(s.Args) > 0 {
			fmt.Fprintf(w, " %s", s.Args)
		}
		fmt.Fprintf(w, "\n%s\n\n", s.Script)
	}
}

func writeFilesIfExists(w io.Writer, f []Files) {
	for _, f := range f {
		if len(f.Files) > 0 {
			fmt.Fprintf(w, "%%files")
			if len(f.Args) > 0 {
				fmt.Fprintf(w, " %s", f.Args)
			}
			fmt.Fprintln(w)

			for _, ft := range f.Files {
				fmt.Fprintf(w, "\t%s\t%s\n", ft.Src, ft.Dst)
			}
			fmt.Fprintln(w)
		}
	}
}

func writeLabelsIfExists(w io.Writer, l map[string]string) {
	if len(l) > 0 {
		fmt.Fprintln(w, "%labels")
		for k, v := range l {
			fmt.Fprintf(w, "\t%s %s\n", k, v)
		}
		fmt.Fprintln(w)
	}
}

// populateRaw is a helper func to output a Definition struct
// into a definition file.
func populateRaw(d *Definition, w io.Writer) {
	// ensure bootstrap is the first parameter in the header
	if v, ok := d.Header["bootstrap"]; ok {
		fmt.Fprintf(w, "%s: %s\n", "bootstrap", v)
	}

	for k, v := range d.Header {
		// filter out bootstrap parameter since it should already be added
		if k == "bootstrap" {
			continue
		}

		fmt.Fprintf(w, "%s: %s\n", k, v)
	}
	fmt.Fprintln(w)

	writeLabelsIfExists(w, d.ImageData.Labels)
	writeFilesIfExists(w, d.BuildData.Files)

	writeSectionIfExists(w, "help", d.ImageData.Help)
	writeSectionIfExists(w, "environment", d.ImageData.Environment)
	writeSectionIfExists(w, "runscript", d.ImageData.Runscript)
	writeSectionIfExists(w, "test", d.ImageData.Test)
	writeSectionIfExists(w, "startscript", d.ImageData.Startscript)
	writeSectionIfExists(w, "pre", d.BuildData.Pre)
	writeSectionIfExists(w, "setup", d.BuildData.Setup)
	writeSectionIfExists(w, "post", d.BuildData.Post)
}
