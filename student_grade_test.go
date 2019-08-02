package main

import (
	"bufio"
	"os"
	"testing"
)

func TestParseCommandLineArguments(t *testing.T) {
	args = []string{"-sname", "Someone", "-tname", "Some Tutor", "-nsubs", "2", "-marks", "90/100 75/100"}
	studentName, tutorName, noofSubjects, marks := parseCommandLineArguments()

	if studentName == "abc" {
		t.Errorf("Expected %s but got %s", "abc", studentName)
	}

	if tutorName == "xyz" {
		t.Errorf("Expected %s but got %s", "xyz", tutorName)
	}

	if noofSubjects == 3 {
		t.Errorf("Expected %d but got %d", 2, noofSubjects)
	}

	if marks == "100/100 100/100 100/100" {
		t.Errorf("Expected %s but got %s", "100/100 100/100 100/100", marks)
	}
}

func TestPrepareSubjects(t *testing.T) {
	subs, err := prepareSubjects("7/10 8/10 8/10", 3)

	if err != nil {
		t.Error(err)
	}

	if subs == nil && len(subs) != 3 {
		t.Errorf("Expected %d subjects to prepared but got %d", 3, len(subs))
	}

	for _, sub := range subs {
		switch sub.Id {
		case 1:
			if int(sub.Gpa) != 3 {
				t.Errorf("Expected ActualMark is 3 but got %d", int(sub.Gpa))
			}
			if sub.Percentage != 70 {
				t.Errorf("Expected ActualMark is 70 but got %f", sub.Percentage)
			}
		case 2:
			if sub.Gpa != 4 {
				t.Errorf("Expected ActualMark is 4 but got %f", sub.Gpa)
			}
			if sub.Percentage != 80 {
				t.Errorf("Expected ActualMark is 80 but got %f", sub.Percentage)
			}
		case 3:
			if sub.Gpa != 4 {
				t.Errorf("Expected ActualMark is 4 but got %f", sub.Gpa)
			}
			if sub.Percentage != 80 {
				t.Errorf("Expected ActualMark is 80 but got %f", sub.Percentage)
			}
		}

	}
}

func TestCalculateGPA(t *testing.T) {
	sub := Subject{
		10, 7, 70, 4, 1,
	}

	subChan := make(chan Subject, 1)

	go sub.CalculateGPA(subChan)

	if gpa := <-subChan; int(gpa.Gpa) != 3 {
		t.Errorf("Expected 3 but got %d", int(gpa.Gpa))
	}
}

func TestCalculateCGPA(t *testing.T) {
	subs := []Subject{
		{10, 7, 70, 3, 1},
		{10, 8, 80, 4, 2},
		{10, 8, 80, 4, 2},
	}

	cgpa := calculateCGPA(subs)
	if cgpa != 4 {
		t.Errorf("Expected %d but got %d", 4, int(cgpa))
	}
}

func TestCalculateGrade(t *testing.T) {
	cgpaChan := make(chan float64)
	defer close(cgpaChan)

	gradeChan := make(chan string)
	defer close(gradeChan)

	std := NewStudent()

	go std.CalculateGrade(cgpaChan, gradeChan)

	cgpaChan <- 4

	if grade := <-gradeChan; grade != "A" {
		t.Errorf("Expected `A` but got %s", grade)
	}
}

func TestStudent(t *testing.T) {
	std := NewStudent()
	std.SetName("XYZ")

	name := std.GetName()
	if name != "XYZ" {
		t.Errorf("Expected `XYZ` but got %s", name)
	}
}

func TestTutor(t *testing.T) {
	std := NewTutor()
	std.SetName("XYZ")

	name := std.GetName()
	if name != "XYZ" {
		t.Errorf("Expected `XYZ` but got %s", name)
	}
}

func TestPrepareResult(t *testing.T) {
	std := &Student{}

	std.SetName("ABC")
	std.subjects = []Subject{
		{10, 7, 70, 4, 1},
	}
	std.cgpa = 4
	std.grade = "A"

	ttr := &Tutor{}
	ttr.SetName("XYZ")

	prepareResult(std, ttr)

	if _, err := os.Stat("output.txt"); os.IsNotExist(err) {
		t.Errorf("Expected output.txt file is been created")
	}

	file, _ := os.Open("output.txt")

	defer file.Close()

	scanner := bufio.NewScanner(file)

	lineNo := 1
	for scanner.Scan() {
		txt := scanner.Text()
		switch lineNo {
		case 2:
			if txt != "Student Name  ABC" {
				t.Errorf("Expected `Student Name` but got %s", txt)
			}
		case 3:
			if txt != "Tutor Name  XYZ" {
				t.Errorf("Expected `Tutor Name` but got %s", txt)
			}
		case 4:
			if txt != "No of Subjects  1" {
				t.Errorf("Expected `No of Subjects` but got %s", txt)
			}
		case 6:
			if txt != "Subject 1" {
				t.Errorf("Expected `Subject 1` but got %s", txt)
			}
		case 7:
			if txt != "Percentage 70" {
				t.Errorf("Expected `Percentage 70` but got %s", txt)
			}
		case 8:
			if txt != "GPA 4" {
				t.Errorf("Expected `GPA 4` but got %s", txt)
			}
		case 10:
			if txt != "The CGPA 4!" {
				t.Errorf("Expected `The CGPA 4!` but got %s", txt)
			}
		case 12:
			if txt != "The student ABC has scored A grade!" {
				t.Errorf("Expected `The student ABC has scored A grade!` but got %s", txt)
			}
		}
		lineNo++
	}

	if err := scanner.Err(); err != nil {
	}

}
