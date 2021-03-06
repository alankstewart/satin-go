package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"time"
)

const (
	N      = 10
	RAD    = 18E-2
	W1     = 3E-1
	DR     = 2E-3
	DZ     = 4E-2
	LAMBDA = 10.6E-3
	AREA   = math.Pi * (RAD * RAD)
	Z1     = math.Pi * (W1 * W1) / LAMBDA
	Z12    = Z1 * Z1
	EXPR   = 2 * math.Pi * DR
	INCR   = 8001
)

type gaussian struct {
	inputPower          int
	saturationIntensity int
	outputPower         float64
}

type laser struct {
	smallSignalGain   float32
	dischargePressure int
	outputFile        string
	carbonDioxide     string
}

var ci chan int = nil

func main() {
	concurrent := flag.Bool("concurrent", true, "Run via gorountines")
	flag.Parse()

	start := time.Now()
	Calculate(*concurrent)
	end := time.Now()
	fmt.Printf("The time was %v.\n", end.Sub(start))
}

func Calculate(concurrent bool) {
	inputPowers := getInputPowers() // immutable; shared by goroutines
	laserData := getLaserData()     // immutable; shared by goroutines
	lNum := len(laserData)

	if concurrent {
		runtime.GOMAXPROCS(runtime.NumCPU())
		ci = make(chan int, lNum)
	}

	for l := 0; l < lNum; l++ {
		if ci == nil {
			process(inputPowers, laserData[l])
		} else {
			go process(inputPowers, laserData[l])
		}
	}

	if ci != nil {
		i := 0
	L:
		for {
			select {
			case <-ci:
				i++
				if i == lNum {
					break L
				}
			}
		}
	}
}

func getInputPowers() []int {
	const inputPowerFile = "pin.dat"
	fd, err := os.Open(inputPowerFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	inputPowers := make([]int, N)
	i := 0
	for {
		_, err := fmt.Fscanf(fd, "%d \n", &inputPowers[i])
		if err != nil && err == io.EOF {
			inputPowers = inputPowers[0:i]
			return inputPowers
		}

		i++
		if i == len(inputPowers) {
			newSlice := make([]int, i*2)
			copy(newSlice, inputPowers)
			inputPowers = newSlice
		}
	}
	return nil
}

func getLaserData() []laser {
	const laserDataFile = "laser.dat"
	fd, err := os.Open(laserDataFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	laserData := make([]laser, N)
	i := 0
	for {
		_, err := fmt.Fscanf(fd, "%s %f %d %s \n", &laserData[i].outputFile, &laserData[i].smallSignalGain, &laserData[i].dischargePressure, &laserData[i].carbonDioxide)
		if err != nil && err == io.EOF {
			laserData = laserData[0:i]
			return laserData
		}

		i++
		if i == len(laserData) {
			newSlice := make([]laser, i*2)
			copy(newSlice, laserData)
			laserData = newSlice
		}
	}
	return nil
}

func process(inputPowers []int, laserData laser) {
	fd, err := os.Create(laserData.outputFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	fmt.Fprintf(fd, "Start date: %s\n\nGaussian Beam\n\nPressure in Main Discharge = %dkPa\nSmall-signal Gain = %4.1f\nCO2 via %s\n\nPin\t\tPout\t\tSat. Int\tln(Pout/Pin)\tPout-Pin\n(watts)\t\t(watts)\t\t(watts/cm2)\t\t\t(watts)\n",
		time.Now(), laserData.dischargePressure, laserData.smallSignalGain, laserData.carbonDioxide)

	pNum := len(inputPowers)
	for i := 0; i < pNum; i++ {
		var gaussianData = new([16]gaussian)
		gaussianCalculation(inputPowers[i], laserData.smallSignalGain, gaussianData)
		for j := 0; j < len(gaussianData); j++ {
			inputPower := gaussianData[j].inputPower
			outputPower := gaussianData[j].outputPower
			fmt.Fprintf(fd, "%d\t\t%7.3f\t\t%d\t\t%5.3f\t\t%7.3f\n", inputPower, outputPower, gaussianData[j].saturationIntensity, math.Log(outputPower/float64(inputPower)), outputPower-float64(inputPower))
		}
	}

	fmt.Fprintf(fd, "\nEnd date: %s\n", time.Now())
	if ci != nil {
		ci <- pNum
	}
	return
}

func gaussianCalculation(inputPower int, smallSignalGain float32, gaussianData *[16]gaussian) {
	var expr1 = new([INCR]float64)

	for i := 0; i < INCR; i++ {
		zInc := (float64(i) - INCR/2) / 25
		expr1[i] = zInc * 2 * DZ / (Z12 + math.Pow(zInc, 2))
	}

	inputIntensity := 2 * float64(inputPower) / AREA
	expr2 := float64((smallSignalGain / 32E3) * DZ)

	i := 0
	var waitChan chan bool
	if ci != nil {
		waitChan = make(chan bool, 15)
	}
	for saturationIntensity := 10E3; saturationIntensity <= 25E3; saturationIntensity += 1E3 {
		gaussians := &gaussianData[i]
		if ci == nil {
			gcalc(inputPower, expr1, inputIntensity, expr2, saturationIntensity, gaussians, nil)
		} else {
			go gcalc(inputPower, expr1, inputIntensity, expr2, saturationIntensity, gaussians, waitChan)
		}
		i++
	}
	if ci != nil {
		for saturationIntensity := 10E3; saturationIntensity <= 25E3; saturationIntensity += 1E3 {
			<-waitChan
		}
	}
}

func gcalc(inputPower int, expr1 *[INCR]float64, inputIntensity float64, expr2 float64, saturationIntensity float64, gaussians *gaussian, waitChan chan bool) {
	outputPower := 0.0
	expr3 := saturationIntensity * expr2

	for r := 0.0; r <= 0.5; r += DR {
		outputIntensity := inputIntensity * math.Exp(-2*math.Pow(r, 2)/math.Pow(RAD, 2))
		for j := 0; j < INCR; j++ {
			outputIntensity *= (1 + expr3/(saturationIntensity+outputIntensity) - expr1[j])
		}
		outputPower += (outputIntensity * EXPR * r)
	}
	gaussians.inputPower = inputPower
	gaussians.saturationIntensity = int(saturationIntensity)
	gaussians.outputPower = outputPower
	if waitChan != nil {
		waitChan <- true
	}
}
