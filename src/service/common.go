package service

func GetSkip(page int, pageSize int) (int, int) {
	skip := 0
	if page < 1 || pageSize < 1 {
		skip = 0
	}
	skip = (page - 1) * pageSize
	return skip, pageSize
}
