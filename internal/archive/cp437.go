package archive

import (
	"strings"
	"unicode/utf8"
)

// cp437High maps bytes 0x80-0xFF to their CP437 runes.
var cp437High = [128]rune{
	'Ç', 'ü', 'é', 'â', 'ä', 'à', 'å', 'ç', 'ê', 'ë', 'è', 'ï', 'î', 'ì', 'Ä', 'Å',
	'É', 'æ', 'Æ', 'ô', 'ö', 'ò', 'û', 'ù', 'ÿ', 'Ö', 'Ü', '¢', '£', '¥', '₧', 'ƒ',
	'á', 'í', 'ó', 'ú', 'ñ', 'Ñ', 'ª', 'º', '¿', '⌐', '¬', '½', '¼', '¡', '«', '»',
	'░', '▒', '▓', '│', '┤', '╡', '╢', '╖', '╕', '╣', '║', '╗', '╝', '╜', '╛', '┐',
	'└', '┴', '┬', '├', '─', '┼', '╞', '╟', '╚', '╔', '╩', '╦', '╠', '═', '╬', '╧',
	'╨', '╤', '╥', '╙', '╘', '╒', '╓', '╫', '╪', '┘', '┌', '█', '▄', '▌', '▐', '▀',
	'α', 'ß', 'Γ', 'π', 'Σ', 'σ', 'µ', 'τ', 'Φ', 'Θ', 'Ω', 'δ', '∞', 'φ', 'ε', '∩',
	'≡', '±', '≥', '≤', '⌠', '⌡', '÷', '≈', '°', '∙', '·', '√', 'ⁿ', '²', '■', ' ',
}

// decodeEntryName returns a UTF-8 name for a zip entry. Entry names that are
// already valid UTF-8 are returned unchanged; the rest are decoded as CP437,
// the encoding the zip spec assigns to names stored without the UTF-8 flag.
// Without this, raw high bytes reach the filesystem and macOS rejects them with
// "illegal byte sequence".
func decodeEntryName(name string) string {
	if utf8.ValidString(name) {
		return name
	}

	var decoded strings.Builder
	decoded.Grow(len(name))
	for i := 0; i < len(name); i++ {
		if b := name[i]; b < 0x80 {
			decoded.WriteByte(b)
		} else {
			decoded.WriteRune(cp437High[b-0x80])
		}
	}
	return decoded.String()
}
