package entity

func TrimLower(s string) string {
	return toLower(trim(s))
}

func trim(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' {
			start := i
			for j := len(s) - 1; j >= start; j-- {
				if s[j] != ' ' && s[j] != '\t' && s[j] != '\n' {
					return toLower(s[start : j+1])
				}
			}
		}
	}
	return ""
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}
