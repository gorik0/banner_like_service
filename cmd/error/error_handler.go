package error

import "log"

func HandleErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s ::: %s", msg, err)
		return
	}
	return
}
