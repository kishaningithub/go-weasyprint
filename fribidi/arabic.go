package fribidi

/* shapeArabic - do Arabic shaping
 *
 * The actual shaping that is done depends on the flags set.  Only flags
 * starting with FRIBIDI_FLAG_SHAPE_ARAB_ affect this function.
 * Currently these are:
 *
 *	* FRIBIDI_FLAG_SHAPE_MIRRORING: Do mirroring.
 *	* FRIBIDI_FLAG_SHAPE_ARAB_PRES: Shape Arabic characters to their
 *					presentation form glyphs.
 *	* FRIBIDI_FLAG_SHAPE_ARAB_LIGA: Form mandatory Arabic ligatures.
 *	* FRIBIDI_FLAG_SHAPE_ARAB_CONSOLE: Perform additional Arabic shaping
 *					   suitable for text rendered on
 *					   grid terminals with no mark
 *					   rendering capabilities.
 *
 * Of the above, FRIBIDI_FLAG_SHAPE_ARAB_CONSOLE is only used in special
 * cases, but the rest are recommended in any environment that doesn't have
 * other means for doing Arabic shaping.  The set of extra flags that enable
 * this level of Arabic support has a shortcut named FRIBIDI_FLAGS_ARABIC.
 */
func shapeArabic(flags Options, embedding_levels []Level,
	/* input and output */
	ar_props []JoiningType, str []rune) {

	if len(str) == 0 {
		return
	}

	if flags&ShapeArabPres != 0 {
		shapeArabicJoining(ar_props, str)
	}

	if flags&ShapeArabLiga != 0 {
		shapeArabicLigature(mandatoryLigaTable, embedding_levels, ar_props, str)
	}

	// if flags&FRIBIDI_FLAG_SHAPE_ARAB_CONSOLE != 0 {
	// 	fribidi_shape_arabic_ligature(console_liga_table, embedding_levels, len, ar_props, str)
	// 	fribidi_shape_arabic_joining(FRIBIDI_GET_ARABIC_SHAPE_NSM, len, ar_props, str)
	// }
}

type PairMap struct {
	pair [2]rune
	to   rune
}

func shapeArabicJoining(ar_props []JoiningType, str []rune /* input and output */) {
	for i, ar := range ar_props {
		if ar.isArabShapes() {
			str[i] = getArabicShapePres(str[i], ar.joinShape())
		}
	}
}

func compPairMap(a, b PairMap) int32 {
	if a.pair[0] != b.pair[0] {
		return a.pair[0] - b.pair[0]
	}
	return a.pair[1] - b.pair[1]
}

func binarySearch(key PairMap, base []PairMap) (PairMap, bool) {
	min, max := 0, len(base)-1
	for min <= max {
		mid := (min + max) / 2
		p := base[mid]
		c := compPairMap(key, p)
		if c < 0 {
			max = mid - 1
		} else if c > 0 {
			min = mid + 1
		} else {
			return p, true
		}
	}
	return PairMap{}, false
}

func findPairMatch(table []PairMap, first, second rune) rune {
	x := PairMap{
		pair: [2]rune{first, second},
	}
	if match, ok := binarySearch(x, table); ok {
		return match.to
	}
	return 0
}

/* Char we place for a deleted slot, to delete later */
const charFill = 0xFEFF

func shapeArabicLigature(table []PairMap, embedding_levels []Level,
	/* input and output */
	ar_props []JoiningType, str []rune) {
	// TODO: This doesn't form ligatures for even-level Arabic text. no big problem though. */
	L := len(embedding_levels)
	size := len(table)
	for i := 0; i < L-1; i++ {
		var c rune
		if str[i] >= table[0].pair[0] && str[i] <= table[size-1].pair[0] {
			c = findPairMatch(table, str[i], str[i+1])
		}

		if embedding_levels[i].isRtl() != 0 && embedding_levels[i] == embedding_levels[i+1] && c != 0 {
			str[i] = charFill
			ar_props[i] |= ligatured
			str[i+1] = c
		}
	}
}

var mandatoryLigaTable = []PairMap{
	{pair: [2]rune{0xFEDF, 0xFE82}, to: 0xFEF5},
	{pair: [2]rune{0xFEDF, 0xFE84}, to: 0xFEF7},
	{pair: [2]rune{0xFEDF, 0xFE88}, to: 0xFEF9},
	{pair: [2]rune{0xFEDF, 0xFE8E}, to: 0xFEFB},
	{pair: [2]rune{0xFEE0, 0xFE82}, to: 0xFEF6},
	{pair: [2]rune{0xFEE0, 0xFE84}, to: 0xFEF8},
	{pair: [2]rune{0xFEE0, 0xFE88}, to: 0xFEFA},
	{pair: [2]rune{0xFEE0, 0xFE8E}, to: 0xFEFC},
}
