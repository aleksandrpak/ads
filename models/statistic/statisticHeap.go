package statistic

type minuteStatistic struct {
	unixMinutes int64
	count       int32
}

type statisticHeap []minuteStatistic

func (h statisticHeap) Len() int {
	return len(h)
}

func (h statisticHeap) Less(i, j int) bool {
	return h[i].unixMinutes < h[j].unixMinutes
}

func (h statisticHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *statisticHeap) Push(x interface{}) {
	*h = append(*h, x.(minuteStatistic))
}

func (h *statisticHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
