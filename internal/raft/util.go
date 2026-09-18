package raft

func compareLogRecency(lastLogIndexA int, lastLogTermA int, lastLogIndexB int, lastLogTermB int) (res int) {
	if lastLogTermA != lastLogTermB {
		return lastLogTermA - lastLogTermB
	}

	return lastLogIndexA - lastLogIndexB
}
