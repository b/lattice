package graphical_visualizer

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/gizak/termui"
	"github.com/pivotal-golang/clock"
)

const (
	graphicalRateDelta = 100 * time.Millisecond
)

//go:generate counterfeiter -o fake_graphical_visualizer/fake_graphical_visualizer.go . GraphicalVisualizer
type GraphicalVisualizer interface {
	PrintDistributionChart(rate time.Duration) error
}

type graphicalVisualizer struct {
	appExaminer app_examiner.AppExaminer
	ui          terminal.UI
	clock       clock.Clock
}

func New(appExaminer app_examiner.AppExaminer, ui terminal.UI, clock clock.Clock) *graphicalVisualizer {
	return &graphicalVisualizer{appExaminer, ui, clock}
}

func (gv *graphicalVisualizer) PrintDistributionChart(rate time.Duration) error {

	//Initialize termui
	err := termui.Init()
	if err != nil {
		return errors.New("Unable to initalize terminal graphics mode.")
		//panic(err)
	}
	defer termui.Close()
	if rate <= time.Duration(0) {
		rate = graphicalRateDelta
	}

	termui.UseTheme("helloworld")

	//Initalize some widgets
	p := termui.NewPar("Lattice Visualization")
	if p == nil {
		return errors.New("Error Initializing termui objects NewPar")
	}
	p.Height = 1
	p.Width = 25
	p.TextFgColor = termui.ColorWhite
	p.Border.FgColor = termui.ColorWhite
	p.HasBorder = false

	r := termui.NewPar(fmt.Sprintf("rate:%v", rate))
	if r == nil {
		return errors.New("Error Initializing termui objects NewPar")
	}
	r.Height = 1
	r.Width = 10
	r.TextFgColor = termui.ColorWhite
	r.Border.FgColor = termui.ColorWhite
	r.HasBorder = false

	s := termui.NewPar("hit [+=inc; -=dec; q=quit]")
	if s == nil {
		return errors.New("Error Initializing termui objects NewPar")
	}
	s.Height = 1
	s.Width = 30
	s.TextFgColor = termui.ColorWhite
	s.Border.FgColor = termui.ColorWhite
	s.HasBorder = false

	bg := termui.NewMBarChart()
	if bg == nil {
		return errors.New("Error Initializing termui objects NewMBarChart")
	}
	bg.IsDisplay = false
	bg.Data[0] = []int{0}
	bg.DataLabels = []string{"1[M]"}
	bg.Width = termui.TermWidth() - 10
	bg.Height = termui.TermHeight() - 5
	bg.BarColor[0] = termui.ColorGreen
	bg.BarColor[1] = termui.ColorYellow
	bg.NumColor[0] = termui.ColorRed
	bg.NumColor[1] = termui.ColorRed
	bg.TextColor = termui.ColorWhite
	bg.Border.LabelFgColor = termui.ColorWhite
	bg.Border.Label = "X-Axis: I[R/T]=CellIndex[Total Instance/Running Instance];[M]=Missing;[E]=Empty"
	bg.BarWidth = 10
	bg.BarGap = 1

	//12 column grid system
	termui.Body.AddRows(termui.NewRow(termui.NewCol(12, 5, p)))
	termui.Body.AddRows(termui.NewRow(termui.NewCol(12, 0, bg)))
	termui.Body.AddRows(termui.NewRow(termui.NewCol(6, 0, s), termui.NewCol(6, 5, r)))

	termui.Body.Align()

	termui.Render(termui.Body)

	bg.IsDisplay = true
	evt := termui.EventCh()
	for {
		select {
		case e := <-evt:
			if e.Type == termui.EventKey {
				switch {
				case (e.Ch == 'q' || e.Ch == 'Q'):
					return nil
				case (e.Ch == '+' || e.Ch == '='):
					rate += graphicalRateDelta
				case (e.Ch == '_' || e.Ch == '-'):
					rate -= graphicalRateDelta
					if rate <= time.Duration(0) {
						rate = graphicalRateDelta
					}
				}
				r.Text = fmt.Sprintf("rate:%v", rate)
				termui.Render(termui.Body)
			}
			if e.Type == termui.EventResize {
				termui.Body.Width = termui.TermWidth()
				termui.Body.Align()
				termui.Render(termui.Body)
			}
		case <-gv.clock.NewTimer(rate).C():
			err := gv.getProgressBars(bg)
			if err != nil {
				return err
			}
			termui.Render(termui.Body)
		}
	}
	return nil
}

func (gv *graphicalVisualizer) getProgressBars(bg *termui.MBarChart) error {

	var barIntList [2][]int
	var barStringList []string

	var barLabel string
	maxTotal := -1

	cells, err := gv.appExaminer.ListCells()
	if err != nil {
		return err
	}

	for i, cell := range cells {

		if cell.Missing {
			barLabel = fmt.Sprintf("%d[M]", i+1)

		} else if cell.RunningInstances == 0 && cell.ClaimedInstances == 0 && !cell.Missing {
			barLabel = fmt.Sprintf("%d[E]", i+1)
			barIntList[0] = append(barIntList[0], 0)
			barIntList[1] = append(barIntList[1], 0)
		} else {

			total := cell.RunningInstances + cell.ClaimedInstances
			barIntList[0] = append(barIntList[0], cell.RunningInstances)
			barIntList[1] = append(barIntList[1], cell.ClaimedInstances)
			barLabel = fmt.Sprintf("%d[%d/%d]", i+1, cell.RunningInstances, total)
			if total > maxTotal {
				maxTotal = total
			}
		}
		barStringList = append(barStringList, barLabel)
	}

	bg.Data[0] = barIntList[0]
	bg.Data[1] = barIntList[1]
	bg.DataLabels = barStringList
	bg.SetMax(maxTotal + 10)

	return nil
}
