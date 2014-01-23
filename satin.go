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
	N      = 100
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

func main() {
	concurrent := flag.Bool("concurrent", false, "Run via gorountines")
	flag.Parse()

	if *concurrent {
		runtime.GOMAXPROCS(runtime.NumCPU())
		ci = make(chan int, N)
	}

	start := time.Now()
	if !Calculate() {
		panic("Failed to complete")
	}
	end := time.Now()
	fmt.Printf("The time was %v.\n", end.Sub(start))
}

var ci chan int = nil

func Calculate() (success bool) {
	var laserData = new([N]laser)
	var inputPowers = new([N]int)
	var total int = 0

	pNum := getInputPowers(inputPowers) // immutable; shared by goroutines
	lNum := getLaserData(laserData)     // immutable; shared by goroutines

	for l := 0; l < lNum; l++ {
		if ci != nil {
			go process(l, pNum, inputPowers, laserData)
		} else {
			total += process(l, pNum, inputPowers, laserData)
		}
	}

	if ci != nil {
		i := 0
	L:
		for {
			select {
			case count := <-ci:
				total += count
				i++
				if i == lNum {
					break L
				}
			}
		}
	}

	return total == pNum*lNum
}

func process(i int, pNum int, inputPowers *[N]int, laserData *[N]laser) (count int) {
	var gaussianData = new([16]gaussian)

	fd, err := os.Create(laserData[i].outputFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	fmt.Fprintf(fd, "Start date: %s\n\nGaussian Beam\n\nPressure in Main Discharge = %dkPa\nSmall-signal Gain = %4.1f\nCO2 via %s\n\nPin\t\tPout\t\tSat. Int\tln(Pout/Pin)\tPout-Pin\n(watts)\t\t(watts)\t\t(watts/cm2)\t\t\t(watts)\n",
		time.Now(), laserData[i].dischargePressure, laserData[i].smallSignalGain, laserData[i].carbonDioxide)
	count = 0
	for j := 0; j < pNum; j++ {
		gaussianCalculation(inputPowers[j], laserData[i].smallSignalGain, gaussianData)
		for k := 0; k < len(gaussianData); k++ {
			inputPower := gaussianData[k].inputPower
			outputPower := gaussianData[k].outputPower
			fmt.Fprintf(fd, "%d\t\t%7.3f\t\t%d\t\t%5.3f\t\t%7.3f\n", inputPower, outputPower, gaussianData[k].saturationIntensity, math.Log(outputPower/float64(inputPower)), outputPower-float64(inputPower))
		}
		count++
	}
	fmt.Fprintf(fd, "\nEnd date: %s\n", time.Now())
	if ci != nil {
		ci <- count
	}
	return
}

func getInputPowers(inputPowers *[N]int) int {
	const inputPowerFile = "pin.dat"
	fd, err := os.Open(inputPowerFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	for i := 0; i < N; i++ {
		_, err := fmt.Fscanf(fd, "%d \n", &inputPowers[i])
		if err != nil && err == io.EOF {
			return i
		}
	}
	return 0
}

func getLaserData(laserData *[N]laser) int {
	const laserDataFile = "laser.dat"
	fd, err := os.Open(laserDataFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	var i int
	for i = 0; i < N; i++ {
		_, err := fmt.Fscanf(fd, "%s %f %d %s \n", &laserData[i].outputFile, &laserData[i].smallSignalGain, &laserData[i].dischargePressure, &laserData[i].carbonDioxide)
		if err != nil && err == io.EOF {
			return i
		}
	}
	return 0
}

func gaussianCalculation(inputPower int, smallSignalGain float32, gaussianData *[16]gaussian) {
	var expr1 = new([INCR]float64)

	for i := 0; i < INCR; i++ {
		zInc := (float64(i) - 4000) / 25
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
			gcalc(saturationIntensity, expr1, expr2, inputPower, inputIntensity, gaussians, nil)
		} else {
			go gcalc(saturationIntensity, expr1, expr2, inputPower, inputIntensity, gaussians, waitChan)
		}
		i++
	}
	if ci != nil {
		for saturationIntensity := 10E3; saturationIntensity <= 25E3; saturationIntensity += 1E3 {
			<-waitChan
		}
	}
}

func gcalc(saturationIntensity float64, expr1 *[INCR]float64, expr2 float64, inputPower int, inputIntensity float64, gaussians *gaussian, waitChan chan bool) {
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
