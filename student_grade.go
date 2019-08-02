package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var args []string

// Person The person base struct
type Person struct {
	name string
}

// GetName gets the name of the Person
func (p *Person) GetName() string {
	return p.name
}

// SetName sets the name of the Person
func (p *Person) SetName(name string) {
	p.name = name
}

// Subject The Subject struct to hold subject details
type Subject struct {
	ActualMark, ObtainedMark, Percentage, Gpa float64
	Id                                        int
}

// CalculateGPA Calculates the GPA
func (s *Subject) CalculateGPA(subChan chan<- Subject) {
	sub := Subject{}
	sub.ActualMark = s.ActualMark
	sub.ObtainedMark = s.ObtainedMark
	sub.Percentage = s.ObtainedMark * 100 / s.ActualMark
	sub.Gpa = sub.Percentage * 5 / 100
	sub.Id = s.Id

	subChan <- sub
}

// Tutor The Tutor struct to hold tutor details
type Tutor struct {
	Person
}

// Student the Student struct to hold the student details
type Student struct {
	Person
	subjects []Subject
	cgpa     float64
	grade    string
}

// NewStudent returns new instance of Student struct
func NewStudent() *Student {
	return &Student{Person: Person{}, cgpa: 0, subjects: []Subject{}}
}

// NewTutor returns new instance of Student struct
func NewTutor() *Tutor {
	return &Tutor{Person: Person{}}
}

// CalculateGrade calculates the grade
func (s *Student) CalculateGrade(cgpaChan <-chan float64, gradeChan chan<- string) {
	cgpa := <-cgpaChan
	switch true {
	case cgpa >= 4:
		gradeChan <- "A"
		break
	case cgpa >= 3 && cgpa < 4:
		gradeChan <- "B"
		break
	case cgpa >= 2 && cgpa < 3:
		gradeChan <- "C"
		break
	case cgpa >= 1 && cgpa < 2:
		gradeChan <- "D"
		break
	default:
		gradeChan <- "F"
	}
}

func mustPrepareSubjects(subjects []Subject, err error) []Subject {
	if err != nil {
		panic(err)
	}

	return subjects
}

func prepareSubjects(marks string, noofSubjects int) ([]Subject, error) {
	totalMarksAsString := strings.Split(marks, " ")

	re := regexp.MustCompile(`(?m)(?:\b|-)([1-9]{1,2}[0]?|100)\b\/(?:\b|-)([1-9]{1,2}[0]?|100)\b`)

	subjects := []Subject{}

	subChan := make(chan Subject, noofSubjects)

	for i, mrk := range totalMarksAsString {

		if !re.MatchString(mrk) {
			flag.Usage()
			return nil, errors.New("mark format is wrong")
		}

		totalMarks := strings.Split(mrk, "/")

		obtainedMark, _ := strconv.ParseFloat(totalMarks[0], 10)
		actualMark, _ := strconv.ParseFloat(totalMarks[1], 10)

		sub := Subject{
			Id:           i + 1,
			ActualMark:   actualMark,
			ObtainedMark: obtainedMark}

		go sub.CalculateGPA(subChan)
	}

	for gpaI := 0; gpaI < noofSubjects; gpaI++ {
		gpa := <-subChan
		subjects = append(subjects, gpa)
	}

	return subjects, nil
}

func print(std *Student, ttr *Tutor) {
	fmt.Println("\nStudent Name ", std.GetName())
	fmt.Println("Tutor Name ", ttr.GetName())
	fmt.Println("No of Subjects ", len(std.subjects))

	sort.Slice(std.subjects, func(i, j int) bool {
		return std.subjects[i].Id < std.subjects[j].Id
	})

	for i, sub := range std.subjects {
		fmt.Printf("\nSubject %d\n", i+1)
		fmt.Printf("Percentage %d\n", int(sub.Percentage))
		fmt.Printf("GPA %d\n", int(sub.Gpa))
	}

	fmt.Printf("\nThe CGPA %d!", int(std.cgpa))
	fmt.Printf("\n\nThe student %s has scored %s grade!", std.GetName(), std.grade)
}

func prepareResult(std *Student, ttr *Tutor) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	print(std, ttr)

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC

	fmt.Print(out)

	outfile, err := os.Create("output.txt")

	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	outfile.WriteString(out)
}

func calculateCGPA(subs []Subject) float64 {
	cgpa := 0.0
	for _, sub := range subs {
		cgpa += sub.Gpa
	}

	return math.Ceil(cgpa / float64(len(subs)))
}

func parseCommandLineArguments() (string, string, int, string) {
	studentName := flag.String("sname", "abc", "Name of the student")
	tutorName := flag.String("tname", "xyz", "Name of the tutor")
	noofSubjects := flag.Int("nsubs", 3, "No of subjects for the given semester")
	marks := flag.String("marks", "100/100 100/100 100/100", "Total marks in the semester")

	a := os.Args[1:]
	if args != nil {
		a = args
	}
	flag.CommandLine.Parse(a)

	return *studentName, *tutorName, *noofSubjects, *marks
}

func main() {
	studentName, tutorName, noofSubjects, marks := parseCommandLineArguments()

	std := NewStudent()
	std.SetName(studentName)

	ttr := NewTutor()
	ttr.SetName(tutorName)

	std.subjects = mustPrepareSubjects(prepareSubjects(marks, noofSubjects))

	if len(std.subjects) != noofSubjects {
		flag.Usage()
		panic("No of subjects given as arguments is not matching with the marks provided!")
	}

	std.cgpa = calculateCGPA(std.subjects)

	cgpaChan := make(chan float64)
	defer close(cgpaChan)
	gradeChan := make(chan string)
	defer close(gradeChan)

	go std.CalculateGrade(cgpaChan, gradeChan)

	cgpaChan <- std.cgpa
	std.grade = <-gradeChan

	prepareResult(std, ttr)
}
