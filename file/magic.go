package main

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -lmagic -L/usr/local/lib
#include <stdlib.h>
#include <magic.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

var (
	ConnectionFailure = errors.New("libmagic: Failed to open magic database.")
	ConnectionError   = errors.New("libmagic: Connection already closed.")
	ATimeError        = errors.New("libmagic: MAGIC_PRESERVE_ATIME unsupported")
)

type Flag int

const (
	// No special handling
	MAGIC_NONE Flag = C.MAGIC_NONE

	// Prints debugging messages to stderr
	MAGIC_DEBUG Flag = C.MAGIC_DEBUG

	// If the file queried is a symlink, follow it.
	MAGIC_SYMLINK Flag = C.MAGIC_SYMLINK

	// If the file is compressed, unpack it and look at the contents.
	MAGIC_COMPRESS Flag = C.MAGIC_COMPRESS

	// If the file is a block or character special device, then open the device
	// and try to look in its contents.
	MAGIC_DEVICES Flag = C.MAGIC_DEVICES

	// Return a MIME type string, instead of a textual description.
	MAGIC_MIME_TYPE Flag = C.MAGIC_MIME_TYPE

	// Return a MIME encoding, instead of a textual description.
	MAGIC_MIME_ENCODING Flag = C.MAGIC_MIME_ENCODING

	// A shorthand for MAGIC_MIME_TYPE | MAGIC_MIME_ENCODING.
	MAGIC_MIME Flag = C.MAGIC_MIME

	// Return all matches, not just the first.
	MAGIC_CONTINUE Flag = C.MAGIC_CONTINUE

	// Check the magic database for consistency and print warnings to stderr.
	MAGIC_CHECK Flag = C.MAGIC_CHECK

	// On systems that support utime(2) or utimes(2), attempt to preserve the
	// access time of files analyzed.
	MAGIC_PRESERVE_ATIME Flag = C.MAGIC_PRESERVE_ATIME

	// Don't translate unprintable characters to a \ooo octal representation.
	MAGIC_RAW Flag = C.MAGIC_RAW

	// Treat operating system errors while trying to open files and follow
	// symlinks as real errors, instead of printing them in the magic buffer
	MAGIC_ERROR Flag = C.MAGIC_ERROR

	// Return the Apple creator and type.
	MAGIC_APPLE Flag = C.MAGIC_APPLE

	// Don't check for EMX application type (only on EMX).
	MAGIC_NO_CHECK_APPTYPE Flag = C.MAGIC_NO_CHECK_APPTYPE

	// Don't get extra information on MS Composite Document Files.
	MAGIC_NO_CHECK_CDF Flag = C.MAGIC_NO_CHECK_CDF

	// Don't look inside compressed files.
	MAGIC_NO_CHECK_COMPRESS Flag = C.MAGIC_NO_CHECK_COMPRESS

	// Don't print ELF details.
	MAGIC_NO_CHECK_ELF Flag = C.MAGIC_NO_CHECK_ELF

	// Don't check text encodings.
	MAGIC_NO_CHECK_ENCODING Flag = C.MAGIC_NO_CHECK_ENCODING

	// Don't consult magic files.
	MAGIC_NO_CHECK_SOFT Flag = C.MAGIC_NO_CHECK_SOFT

	// Don't examine tar files.
	MAGIC_NO_CHECK_TAR Flag = C.MAGIC_NO_CHECK_TAR

	// Don't check for various types of text files.
	MAGIC_NO_CHECK_TEXT Flag = C.MAGIC_NO_CHECK_TEXT

	// Don't look for known tokens inside ascii files.
	MAGIC_NO_CHECK_TOKENS Flag = C.MAGIC_NO_CHECK_TOKENS
)

type Magic struct {
	// libmagic database descriptor
	cookie C.magic_t
}

// Open a default instance of libmagic
func Open(flags Flag) (*Magic, error) {
	cookie := C.magic_open(C.int(0))
	if cookie == nil {
		return nil, ConnectionFailure
	}

	m := &Magic{cookie}

	if err := m.SetFlags(flags); err != nil {
		return nil, err
	}

	if err := m.Load(""); err != nil {
		return nil, err
	}

	return m, nil
}

// Close and cleanup libmagic instance
func (m *Magic) Close() (err error) {
	if m.cookie == nil {
		return ConnectionError
	}

	C.magic_close(m.cookie)
	m.cookie = nil
	return
}

// Get a textual representation of the last error
func (m *Magic) check() error {

	err := C.magic_error(m.cookie)
	if err == nil {
		return nil
	}

	return errors.New(C.GoString(err))
}

// Set flags for a libmagic instance
func (m *Magic) SetFlags(flags Flag) error {
	if m.cookie == nil {
		return ConnectionError
	}

	if C.magic_setflags(m.cookie, C.int(flags)) < 0 {
		return ATimeError
	}

	return m.check()
}

// Choose libmagic file to use, to use the system default
// pass an empty string of ""
func (m *Magic) Load(filename string) error {
	if m.cookie == nil {
		return ConnectionError
	}

	if filename == "" {
		C.magic_load(m.cookie, nil)
	} else {
		var cfilename *C.char
		cfilename = C.CString(filename)
		defer C.free(unsafe.Pointer(cfilename))
		C.magic_load(m.cookie, cfilename)
	}

	return m.check()
}

// Use a libmagic instance to identify a file on the disk
func (m *Magic) File(filename string) (string, error) {
	if m.cookie == nil {
		return "", ConnectionError
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	out := C.magic_file(m.cookie, cfilename)
	if out == nil {
		return "", m.check()
	}

	return C.GoString(out), nil
}

// Use a byte array to identify a file
func (m *Magic) Buffer(binary []byte) (string, error) {
	bytes := unsafe.Pointer(&binary[0])
	out := C.magic_buffer(m.cookie, bytes, C.size_t(len(binary)))
	if out == nil {
		return "", m.check()
	}
	return C.GoString(out), nil
}
