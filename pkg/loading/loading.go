package loading

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type Service struct {
	name    string
	percent int
}

type process struct {
	items       []Service
	chanProcess chan<- Service
	cnt         int
	percent     int
	wg          *sync.WaitGroup
}

func NewLoading(packages []string, wg *sync.WaitGroup) *process {
	process := new(process)
	process.cnt = len(packages)
	process.wg = wg
	process.chanProcess = listeners(process)
	process.percent = 0

	return process
}

func listeners(process *process) chan<- Service {
	ch := make(chan Service)

	go func() {
		for service := range ch {
			process.items = append(process.items, service)
			process.print()
			process.wg.Done()
		}
	}()

	return ch
}

func (p *process) Update(service_name string, percent int) {
	p.wg.Add(1)

	go func() {
		p.chanProcess <- Service{
			name:    service_name,
			percent: percent,
		}
	}()
}

func (p *process) print() {
	cmdy := exec.Command("clear") //Linux example, its tested
	cmdy.Stdout = os.Stdout
	cmdy.Run()

	for _, service := range p.items {
		fmt.Print(service.name, ":[")
		for j := 0; j <= service.percent; j++ {
			fmt.Print("=")
		}
		for j := service.percent + 1; j <= 100; j++ {
			fmt.Print(" ")
		}
		fmt.Print("] %", service.percent)
		fmt.Println()
	}

}

func (p *process) updatePercent() {
	sum := 0
	cnt := 0
	for _, service := range p.items {
		cnt++
		sum += service.percent
	}

	if cnt == 0 {
		p.percent = 0
		return
	}

	p.percent = int(sum / cnt)
}

func (p *process) Close() {
	close(p.chanProcess)
}
