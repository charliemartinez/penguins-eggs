package cmd

import (
	"os"
	"strings"

	"coa/pkg/parser"
)

const (
	// Il Primary Volume Descriptor ISO9660 sta nel settore 16 (2048 byte/settore).
	isoPVDOffset = 16 * 2048
	// Dentro il PVD il Volume Identifier occupa i byte 40..71.
	isoVolIDStart = 40
	isoVolIDEnd   = 72
)

// readISOVolumeLabel restituisce il Volume Identifier di una ISO9660 leggendo
// direttamente il Primary Volume Descriptor. Se il file non e' una ISO leggibile
// o non ha label, ricade su parser.DefaultLiveLabel.
func readISOVolumeLabel(isoPath string) string {
	f, err := os.Open(isoPath)
	if err != nil {
		return parser.DefaultLiveLabel
	}
	defer f.Close()

	buf := make([]byte, isoVolIDEnd)
	if _, err := f.ReadAt(buf, isoPVDOffset); err != nil {
		return parser.DefaultLiveLabel
	}
	// type 1 = PVD, identificatore standard "CD001"
	if buf[0] != 1 || string(buf[1:6]) != "CD001" {
		return parser.DefaultLiveLabel
	}

	label := strings.TrimRight(string(buf[isoVolIDStart:isoVolIDEnd]), " \x00")
	if label == "" {
		return parser.DefaultLiveLabel
	}
	return label
}
