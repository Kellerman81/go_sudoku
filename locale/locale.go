// Package locale provides UI string translations.
package locale

// Lang identifies a supported language.
type Lang int

const (
	EN Lang = iota
	DE
)

var Names = []string{"English", "Deutsch"}

// Strings holds all translatable UI strings.
type Strings struct {
	// Menu
	AppTitle      string
	GameType      string
	Difficulty    string
	MaxHints      string
	MaxErrors     string
	NewGame       string
	History       string
	ExportPDF     string
	Generating    string
	Language      string

	// Game types / difficulties
	Normal string
	Killer string
	Samurai string
	Easy   string
	Medium string
	Hard   string
	Expert string

	// Side panel
	NotesOff    string
	NotesOn     string
	Hint        string
	HintLeft    string // format: "Hint (%d left)"
	HintNone    string
	Undo        string
	Check       string
	GiveUp      string
	Menu        string
	ErrorsCount string // format: "Errors: %d"
	ErrorsLimit string // format: "Errors: %d / %d"
	HintsCount  string // format: "Hints: %d"
	HintsLimit  string // format: "Hints: %d / %d"

	// Confirm dialogs
	GiveUpMsg   string
	AbandonMsg  string
	Cancel      string
	OK          string

	// Game over
	YouWon      string
	GameOver    string
	WonSub      string // format: "Solved in %s with %d hint(s) and %d error(s)."
	LostSub     string // format: "You made %d error(s). Better luck next time!"
	PlayAgain   string
	BackToMenu  string

	// History
	HistoryTitle string
	GamesPlayed  string
	Won          string
	GaveUp       string
	TotalHints   string
	BestTime     string
	AvgWinTime   string
	AvgPoints    string
	MaxPoints    string
	Clear        string
	Back         string
	Abandoned    string

	// PDF
	PDFTitle        string
	NormalPuzzles   string
	KillerPuzzles   string
	SamuraiPuzzles  string
	PuzzlesPerPage  string
	SolutionsOff    string
	SolutionsOn     string
	HistoryOff      string
	HistoryOn       string
	GeneratePDF     string
	GeneratingPDF   string
	PDFSaved        string // format: "Saved: %s"

	// About
	About           string
	AboutTitle      string
	AboutDesc       string
	AboutFeatures   string
	AboutFeatureList string
	AboutSource     string
	AboutTech       string
	AboutTechList   string
}

var translations = map[Lang]Strings{
	EN: {
		AppTitle:      "Go Sudoku",
		GameType:      "GAME TYPE",
		Difficulty:    "DIFFICULTY",
		MaxHints:      "MAX HINTS",
		MaxErrors:     "MAX ERRORS",
		NewGame:       "▶  New Game",
		History:       "📋  History",
		ExportPDF:     "📄  Export PDF",
		Generating:    "Generating puzzle…",
		Language:      "LANGUAGE",

		Normal:  "Normal",
		Killer:  "Killer",
		Samurai: "Samurai",
		Easy:    "Easy",
		Medium:  "Medium",
		Hard:    "Hard",
		Expert:  "Expert",

		NotesOff:    "Notes: OFF",
		NotesOn:     "Notes: ON",
		Hint:        "Hint",
		HintLeft:    "Hint (%d left)",
		HintNone:    "Hint (0 left)",
		Undo:        "Undo",
		Check:       "Check",
		GiveUp:      "Give Up",
		Menu:        "Menu",
		ErrorsCount: "Errors: %d",
		ErrorsLimit: "Errors: %d / %d",
		HintsCount:  "Hints: %d",
		HintsLimit:  "Hints: %d / %d",

		GiveUpMsg:  "Give up and reveal the solution?",
		AbandonMsg: "Abandon game and return to menu?",
		Cancel:     "Cancel",
		OK:         "OK",

		YouWon:     "You Won!",
		GameOver:   "Game Over",
		WonSub:     "Solved in %s with %d hint(s) and %d error(s).",
		LostSub:    "You made %d error(s). Better luck next time!",
		PlayAgain:  "Play Again",
		BackToMenu: "Menu",

		HistoryTitle: "Game History",
		GamesPlayed:  "Games Played",
		Won:          "Won",
		GaveUp:       "Gave Up",
		TotalHints:   "Total Hints",
		BestTime:     "Best Time",
		AvgWinTime:   "Avg Win Time",
		AvgPoints:    "Avg Score",
		MaxPoints:    "Best Score",
		Clear:        "Clear",
		Back:         "← Back",
		Abandoned:    "Abandoned",

		PDFTitle:       "Export PDF Book",
		NormalPuzzles:  "NORMAL PUZZLES",
		KillerPuzzles:  "KILLER PUZZLES",
		SamuraiPuzzles: "SAMURAI PUZZLES",
		PuzzlesPerPage: "PUZZLES PER PAGE",
		SolutionsOff:   "Solutions: OFF",
		SolutionsOn:    "Solutions: ON",
		HistoryOff:     "History: OFF",
		HistoryOn:      "History: ON",
		GeneratePDF:    "Generate PDF",
		GeneratingPDF:  "Generating PDF…",
		PDFSaved:       "Saved: %s",

		About:            "ℹ  About",
		AboutTitle:       "About",
		AboutDesc:        "A Sudoku puzzle game written in Go.",
		AboutFeatures:    "FEATURES",
		AboutFeatureList: "• Normal, Killer & Samurai Sudoku\n• Configurable hints & error limits\n• Pencil notes mode (press N)\n• Play history & statistics\n• PDF puzzle book export",
		AboutSource:      "SOURCE CODE",
		AboutTech:        "BUILT WITH",
		AboutTechList:    "Go · Gio UI · gofpdf",
	},
	DE: {
		AppTitle:      "Go Sudoku",
		GameType:      "SPIELTYP",
		Difficulty:    "SCHWIERIGKEIT",
		MaxHints:      "MAX. HINWEISE",
		MaxErrors:     "MAX. FEHLER",
		NewGame:       "▶  Neues Spiel",
		History:       "📋  Verlauf",
		ExportPDF:     "📄  PDF exportieren",
		Generating:    "Rätsel wird erstellt…",
		Language:      "SPRACHE",

		Normal:  "Normal",
		Killer:  "Killer",
		Samurai: "Samurai",
		Easy:    "Leicht",
		Medium:  "Mittel",
		Hard:    "Schwer",
		Expert:  "Experte",

		NotesOff:    "Notizen: AUS",
		NotesOn:     "Notizen: EIN",
		Hint:        "Hinweis",
		HintLeft:    "Hinweis (%d übrig)",
		HintNone:    "Hinweis (0 übrig)",
		Undo:        "Rückgängig",
		Check:       "Prüfen",
		GiveUp:      "Aufgeben",
		Menu:        "Menü",
		ErrorsCount: "Fehler: %d",
		ErrorsLimit: "Fehler: %d / %d",
		HintsCount:  "Hinweise: %d",
		HintsLimit:  "Hinweise: %d / %d",

		GiveUpMsg:  "Aufgeben und Lösung anzeigen?",
		AbandonMsg: "Spiel abbrechen und zum Menü zurück?",
		Cancel:     "Abbrechen",
		OK:         "OK",

		YouWon:     "Gewonnen!",
		GameOver:   "Spiel vorbei",
		WonSub:     "Gelöst in %s mit %d Hinweis(en) und %d Fehler(n).",
		LostSub:    "Du hast %d Fehler gemacht. Viel Glück beim nächsten Mal!",
		PlayAgain:  "Nochmal spielen",
		BackToMenu: "Menü",

		HistoryTitle: "Spielverlauf",
		GamesPlayed:  "Gespielte Spiele",
		Won:          "Gewonnen",
		GaveUp:       "Aufgegeben",
		TotalHints:   "Hinweise gesamt",
		BestTime:     "Beste Zeit",
		AvgWinTime:   "Ø Gewinnzeit",
		AvgPoints:    "Ø Punkte",
		MaxPoints:    "Beste Punkte",
		Clear:        "Löschen",
		Back:         "← Zurück",
		Abandoned:    "Abgebrochen",

		PDFTitle:       "PDF-Buch exportieren",
		NormalPuzzles:  "NORMALE RÄTSEL",
		KillerPuzzles:  "KILLER-RÄTSEL",
		SamuraiPuzzles: "SAMURAI-RÄTSEL",
		PuzzlesPerPage: "RÄTSEL PRO SEITE",
		SolutionsOff:   "Lösungen: AUS",
		SolutionsOn:    "Lösungen: EIN",
		HistoryOff:     "Verlauf: AUS",
		HistoryOn:      "Verlauf: EIN",
		GeneratePDF:    "PDF erstellen",
		GeneratingPDF:  "PDF wird erstellt…",
		PDFSaved:       "Gespeichert: %s",

		About:            "ℹ  Über",
		AboutTitle:       "Über",
		AboutDesc:        "Ein Sudoku-Spiel, geschrieben in Go.",
		AboutFeatures:    "FUNKTIONEN",
		AboutFeatureList: "• Normal-, Killer- & Samurai-Sudoku\n• Konfigurierbare Hinweis- & Fehlerlimits\n• Notizmodus (N drücken)\n• Spielverlauf & Statistiken\n• PDF-Rätselbuch-Export",
		AboutSource:      "QUELLCODE",
		AboutTech:        "ERSTELLT MIT",
		AboutTechList:    "Go · Gio UI · gofpdf",
	},
}

// Get returns the string table for the given language.
func Get(l Lang) *Strings {
	s := translations[l]
	return &s
}

// GameTypes returns the localised game type names.
func (s *Strings) GameTypes() []string {
	return []string{s.Normal, s.Killer, s.Samurai}
}

// Difficulties returns the localised difficulty names.
func (s *Strings) Difficulties() []string {
	return []string{s.Easy, s.Medium, s.Hard, s.Expert}
}
