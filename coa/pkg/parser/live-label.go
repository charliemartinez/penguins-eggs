package parser

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultLiveLabel e' il volume label usato quando nessun branding del
// wardrobe e' attivo: mantiene il comportamento storico di penguins-eggs.
const DefaultLiveLabel = "OA_LIVE"

// maxLiveLabelLen: il volume id ISO9660 (senza Joliet) e' limitato a 32 byte.
const maxLiveLabelLen = 32

// brandingDescCandidates elenca, in ordine di priorita', i branding.desc
// installati da penguins-tailor quando si indossa un costume del wardrobe.
// E' una variabile (non una costante) per poterla sostituire nei test.
var brandingDescCandidates = []string{
	// <wardrobe>/v2/branding/<nome>/calamares/branding/branding.desc
	"/etc/penguins-eggs.d/branding/calamares/branding/branding.desc",
	// sysroot del costume: etc/penguins-eggs.d/brain.d/assets/calamares/branding.desc
	"/etc/penguins-eggs.d/brain.d/assets/calamares/branding.desc",
}

// brandingNameKeys: chiavi della sezione "strings:" da cui leggere il nome.
var brandingNameKeys = []string{"productName", "shortProductName", "bootloaderEntryName"}

// ResolveLiveLabel restituisce il volume label da assegnare alla ISO live
// (xorriso -volid) e da usare nei parametri di avvio che la cercano per label.
//
// Il nome proviene dal branding del wardrobe attivo (es. "Quirinux"); se non
// c'e' nessun branding leggibile si ricade su DefaultLiveLabel.
func ResolveLiveLabel() string {
	for _, path := range brandingDescCandidates {
		if name := readBrandingProductName(path); name != "" {
			if label := SanitizeLiveLabel(name); label != "" {
				return label
			}
		}
	}
	return DefaultLiveLabel
}

// readBrandingProductName legge il nome della distribuzione da un
// branding.desc di Calamares. Restituisce "" se il file non esiste, non e'
// YAML valido, o il valore e' ancora un placeholder di template ({{ ... }}).
func readBrandingProductName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var desc struct {
		Strings map[string]interface{} `yaml:"strings"`
	}
	if err := yaml.Unmarshal(data, &desc); err != nil {
		return ""
	}

	for _, key := range brandingNameKeys {
		raw, ok := desc.Strings[key]
		if !ok {
			continue
		}
		value, ok := raw.(string)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || strings.Contains(value, "{{") {
			continue
		}
		return value
	}
	return ""
}

// SanitizeLiveLabel trasforma un nome di distribuzione in un volume label
// sicuro: solo [A-Za-z0-9._-], ogni altro carattere (spazi, "/", ecc.)
// diventa "_", massimo 32 caratteri. Il risultato non contiene spazi, quindi
// puo' viaggiare senza quoting nella riga di comando di xorriso, nel kernel
// cmdline e in "search --label" di GRUB. Restituisce "" se non resta nulla di
// utilizzabile.
func SanitizeLiveLabel(name string) string {
	var b strings.Builder
	lastUnderscore := false

	for _, r := range strings.TrimSpace(name) {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '-':
			b.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
		}
	}

	label := strings.Trim(b.String(), "_.-")
	if len(label) > maxLiveLabelLen {
		label = strings.Trim(label[:maxLiveLabelLen], "_.-")
	}
	return label
}
