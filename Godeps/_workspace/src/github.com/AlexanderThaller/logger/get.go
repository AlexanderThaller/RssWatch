package logger

import "strings"

func getParent(lo Logger) (log Logger) {
	// Return root if root
	if lo == defroot {
		log = defroot
		return
	}

	s := strings.Split(string(lo), defseperator)

	// Return root if first level logger
	if len(s) == 1 {
		log = defroot
		return
	}

	// Return root if parent is empty
	if s[0] == "" {
		log = defroot
		return
	}

	l := len(s) - 1
	z := s[0:l]

	log = Logger(strings.Join(z, defseperator))

	return
}

func getPriorityFormat(pr Priority) (col, fom int) {
	switch pr {
	case Debug:
		col = colornone
	case Notice:
		col = colorgreen
	case Info:
		col = colorblue
	case Warning:
		col = coloryellow
	case Error:
		col = coloryellow
		fom = textbold
	case Critical:
		col = colorred
	case Alert:
		col = colorred
		fom = textbold
	case Emergency:
		col = colorred
		fom = textblink
	}

	return
}
