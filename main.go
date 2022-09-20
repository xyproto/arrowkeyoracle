package main

import (
	"math/rand"
	"time"

	"github.com/xyproto/textoutput"
	"github.com/xyproto/vt100"
)

const (
	versionString = "Left or right arrow key predictor 1.0.0"

	escKey        = 27
	qKey          = 113
	rKey          = 114
	arrowLeftKey  = 252
	arrowRightKey = 254
)

func main() {

	o := textoutput.New()

	o.Println("<cyan>-----------------------------------------</cyan>")
	o.Printf("<white> %s </white>\n", versionString)
	o.Println("<cyan>-----------------------------------------</cyan>")
	o.Println()
	o.Println("Try pressing left or right arrow, and see if the computer can predict your keypress.")
	o.Println("Press <white>r</white> to let the computer chose a random left or right arrow keypress.")
	o.Println()
	o.Println("Press <white>q</white> or <white>Esc</white> to quit.")
	o.Println()

	tty, err := vt100.NewTTY()
	if err != nil {
		panic(err)
	}
	tty.SetTimeout(10 * time.Millisecond)

	rand.Seed(time.Now().Unix())

	var (
		lastFive    [5]int
		shufflePush = func(i int) {
			lastFive[0] = lastFive[1]
			lastFive[1] = lastFive[2]
			lastFive[2] = lastFive[3]
			lastFive[3] = lastFive[4]
			lastFive[4] = i
		}
		keyPressed           bool
		statsMap             = make(map[[5]int]int64)
		prediction           string
		correctCounter       int64
		predictionCounter    int64
		leftCount            int64
		rightCount           int64
		maybeExclamationMark string
	)

OUT:
	for {
		keyPressed = true
		switch tty.Key() {
		case arrowLeftKey:
			shufflePush(arrowLeftKey) // left
		case arrowRightKey:
			shufflePush(arrowRightKey) // right
		case rKey:
			if rand.Intn(2) == 0 {
				shufflePush(arrowLeftKey) // left
			} else {
				shufflePush(arrowRightKey) // right
			}
		case escKey, qKey:
			o.Printf("\n<lightblue>Bye%s</lightblue>\n", maybeExclamationMark)
			break OUT
		default:
			keyPressed = false
		}
		if keyPressed {

			o.Print("\nYou pressed: ")
			if lastFive[4] == arrowLeftKey {
				o.Println("<magenta>left</magenta>")
			} else if lastFive[4] == arrowRightKey {
				o.Println("<cyan>right<cyan>")
			}

			switch prediction {
			case "left":
				if lastFive[4] == arrowLeftKey { // CORRECT
					o.Println("<green>The computer correctly predicted</green><gray>:</gray> <magenta>left</magenta>")
					correctCounter++
				} else {
					o.Println("<red>The computer wrongly predicted</red><gray>:</gray> <magenta>left</magenta>")
				}
			case "right":
				if lastFive[4] == arrowRightKey { // CORRECT
					o.Println("<green>The computer correctly predicted</green><gray>:</gray> <cyan>right</cyan>")
					correctCounter++
				} else {
					o.Println("<red>The computer wrongly predicted</red><gray>:</gray> <cyan>right</cyan>")
				}
			default:
				o.Println("<darkgray>No prediction.</darkgray>")
			}
			prediction = ""

			// Increase the counter for the current collection of 5 keypresses
			statsMap[lastFive]++

			// Build a key for what it would look like if left was pushed right now
			var leftKey [5]int
			leftKey[0] = lastFive[1]
			leftKey[1] = lastFive[2]
			leftKey[2] = lastFive[3]
			leftKey[3] = lastFive[4]
			leftKey[4] = arrowLeftKey // left

			// Build a key for what it would look like if right was pushed right now
			var rightKey [5]int
			rightKey[0] = lastFive[1]
			rightKey[1] = lastFive[2]
			rightKey[2] = lastFive[3]
			rightKey[3] = lastFive[4]
			rightKey[4] = arrowRightKey // right

			// Lookup how many times this combination followed by a left has happened
			leftCount = 0
			if count, ok := statsMap[leftKey]; ok {
				leftCount = count
			}

			// Lookup how many times this combination followed by a right has happened
			rightCount = 0
			if count, ok := statsMap[rightKey]; ok {
				rightCount = count
			}

			// Predict the most likely event to happen
			if rightCount >= leftCount {
				prediction = "right"
			} else if rightCount < leftCount {
				prediction = "left"
			}
			predictionCounter++

			// Output stats
			ratio := float64(correctCounter) / float64(predictionCounter)
			maybeExclamationMark = ""
			if ratio > 0.5 {
				maybeExclamationMark = "!"
			}
			o.Printf("The computer is right <yellow>%.2f%%</yellow> of the time%s\n", ratio*100.0, maybeExclamationMark)

			// Output the current keypress and the last 4 keypresses, 5 in total
			o.Print("Current state: ")
			for _, v := range lastFive {
				if v == arrowLeftKey {
					o.Print("<magenta>l</magenta>")
				} else if v == arrowRightKey {
					o.Print("<cyan>r</cyan>")
				}
			}
			o.Println()

		}

	}
	tty.Close()
}
