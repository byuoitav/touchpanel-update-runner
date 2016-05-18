package helpers

import (
	"errors"
	"fmt"
)

// Currently if this changes we need to edit the status updaters and post wait

// TODO: Find a way to make this dynamic (maybe a linked list type thing)
func getTouchpanelStepNames() []string {
	var n []string
	n = append(n,
		"Get IPTable",              // 0
		"CheckCurrentVersion/Date", // 1
		"Remove Old Firmware",      // 2
		"Initialize",               // 3
		"Copy Firmware",            // 4
		"Update Firmware",          // 5
		"Copy Project",             // 6
		"Move Project",             // 7
		"Load Project",             // 8
		"Reload IPTable",           // 9
		"Validate")                 // 10
	return n
}

// CompleteStep sends a complete step to the update channel
func CompleteStep(tp TouchpanelStatus, step int, curStatus string) {
	tp.Steps[step].Completed = true
	tp.CurrentStatus = curStatus

	UpdateChannel <- tp
}

func GetTouchpanelSteps() []step {
	var steps []step

	names := getTouchpanelStepNames()

	for indx := range names {
		steps = append(steps, step{StepName: names[indx], Completed: false})
	}

	return steps
}

// GetCurStatus gets the current step (the first item in the list of steps that isn't completed)
// returns it's name/location in array. Returns an error if completed
func (t *TouchpanelStatus) GetCurrentStep() (int, error) {
	// fmt.Printf("Steps: %v\n", t.Steps)
	for k := range t.Steps {
		if t.Steps[k].Completed == false {
			return k, nil
		}
	}

	return 0, errors.New("Complete")
}

// TODO: Move all steps (0-3) to this paradigm
func EvaluateNextStep(curTP TouchpanelStatus) {
	// -----------------------------------------
	// DEBUG
	// -----------------------------------------
	// for i := 0; i < 7; i++ {
	// 			curTP.Steps[i].Completed = true
	// 	}
	// -----------------------------------------
	// DEBUG
	// -----------------------------------------

	stepIndx, err := curTP.GetCurrentStep()

	if err != nil {
		return
	}

	switch stepIndx { // determine where to go next
	case 0:
		CompleteStep(curTP, stepIndx, "Validating")

		go retrieveIPTable(curTP)
	case 1:
		fmt.Printf("%s We've gotten IP Table.\n", curTP.IPAddress)
		need, str := validateNeed(curTP, false)
		if !need {
			fmt.Printf("%s Not needed: %s\n", curTP.IPAddress, str)
			reportNotNeeded(curTP, "Not Needed: "+str)
			return
		}

		fmt.Printf("%s Done validating.\n", curTP.IPAddress)
		CompleteStep(curTP, stepIndx, "Removing old Firmware")
		go removeOldFirmware(curTP)
	case 2:
		fmt.Printf("%s Old Firmware removed.\n", curTP.IPAddress)
		CompleteStep(curTP, stepIndx, "Initializing")

		go initializeTP(curTP)
	case 3: // Initialize - next is copy firmware
		fmt.Printf("%s Moving to copy firmware.\n", curTP.IPAddress)

		// Set status and update the
		CompleteStep(curTP, stepIndx, "Sending Firmware")
		go sendFirmware(curTP) // ship this off concurrently - don't block
	case 4:
		fmt.Printf("%s Moving to update firmware.\n", curTP.IPAddress)

		CompleteStep(curTP, stepIndx, "Updating Firmware")
		go updateFirmware(curTP)
	case 5:
		fmt.Printf("%s Done updating firmware.\n", curTP.IPAddress)
		CompleteStep(curTP, stepIndx, "Sending Project")

		go copyProject(curTP)
	case 6:
		fmt.Printf("%s Project copied\n", curTP.IPAddress)
		CompleteStep(curTP, stepIndx, "Moving Project")

		go moveProject(curTP)
	case 7:
		fmt.Printf("%s Project Moved\n", curTP.IPAddress)
		CompleteStep(curTP, stepIndx, "Loading Project")

		go loadProject(curTP)
	case 8:
		fmt.Printf("%s Project Loaded\n", curTP.IPAddress)
		CompleteStep(curTP, stepIndx, "Reload IPTable")

		go reloadIPTable(curTP)
	case 9:
		fmt.Printf("%s IPTable loaded\n", curTP.IPAddress)
		CompleteStep(curTP, stepIndx, "Validating")

		go validateTP(curTP)
	default:
	}
}
