package loading

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/pkg/db"
)

type Service struct {
	name    string
	percent int
}

type process struct {
	items       map[string]int
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
	process.items = map[string]int{}

	return process
}

func listeners(process *process) chan<- Service {
	ch := make(chan Service)

	go func() {
		for service := range ch {
			process.items[service.name] = service.percent
			process.print()
			process.updatePercent()

			db.NewBoltDB().Set(service.name, []byte(fmt.Sprintf("%d", service.percent)))
			db.StorePercent(fmt.Sprint(process.percent))

			process.wg.Done()
		}
	}()

	return ch
}

func (p *process) Update(service_name string, percent int) {
	p.wg.Add(1)

	p.chanProcess <- Service{
		name:    service_name,
		percent: percent,
	}
}

func (p *process) print() {
	cmdy := exec.Command("clear") //Linux example, its tested
	cmdy.Stdout = os.Stdout
	cmdy.Run()

	for serviceName, percent := range p.items {
		fmt.Print(serviceName, ":[")
		for j := 0; j <= percent; j++ {
			fmt.Print("=")
		}
		for j := percent + 1; j <= 100; j++ {
			fmt.Print(" ")
		}
		fmt.Print("] %", percent)
		fmt.Println()
	}

}

func (p *process) GetPercent() int {
	return p.percent
}

func (p *process) updatePercent() {
	sum := 0
	cnt := 0
	for _, percent := range p.items {
		cnt++
		sum += percent
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
