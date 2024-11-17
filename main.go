package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type MainTask struct {
	Name              string
	TotalSubtasks     int
	CompletedSubtasks int
	Spinner           spinner.Model
	IsCompleted       bool
}

type model struct {
	tasks            []*MainTask
	mutex            sync.Mutex
	isLoading        bool
	resultTasksAdded bool // NEW: Flag to track if result tasks have been added
}

type subtaskCompleteMsg struct {
	mainTaskName string
}

type mainTaskCompleteMsg struct {
	mainTaskName string
}

type addMainTaskMsg struct {
	mainTask *MainTask
}

func initialModel() *model {
	return &model{
		tasks:            []*MainTask{},
		isLoading:        true,
		mutex:            sync.Mutex{},
		resultTasksAdded: false, // Initialize the flag
	}
}

func (m *model) Init() tea.Cmd {
	return nil // No initial command needed
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	m.mutex.Lock()
	defer m.mutex.Unlock()

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		for _, task := range m.tasks {
			if !task.IsCompleted {
				var cmd tea.Cmd
				task.Spinner, cmd = task.Spinner.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case addMainTaskMsg:
		m.tasks = append(m.tasks, msg.mainTask)
		return m, msg.mainTask.Spinner.Tick

	case subtaskCompleteMsg:
		for _, task := range m.tasks {
			if task.Name == msg.mainTaskName {
				task.CompletedSubtasks++
				if task.CompletedSubtasks >= task.TotalSubtasks {
					task.IsCompleted = true
				}
				break
			}
		}
		if m.allTasksCompleted() {
			if !m.resultTasksAdded {
				// Add result tasks
				m.resultTasksAdded = true
				cmds = append(cmds, m.addResultTasks()...)
			} else {
				m.isLoading = false
			}
		}

	case mainTaskCompleteMsg:
		for _, task := range m.tasks {
			if task.Name == msg.mainTaskName {
				task.IsCompleted = true
				break
			}
		}
		if m.allTasksCompleted() {
			if !m.resultTasksAdded {
				// Add result tasks
				m.resultTasksAdded = true
				cmds = append(cmds, m.addResultTasks()...)
			} else {
				m.isLoading = false
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isLoading {
		return m.summaryView()
	}

	s := "\nTasks Progress:\n\n"

	for _, task := range m.tasks {
		var status string
		if task.IsCompleted {
			status = "✔"
		} else {
			status = task.Spinner.View()
		}

		s += fmt.Sprintf("%s %s [%d/%d]\n", status, task.Name, task.CompletedSubtasks, task.TotalSubtasks)
	}

	s += "\nPress 'q' to quit."
	return s
}

func (m *model) summaryView() string {
	s := "\nAll tasks completed!\n\nSummary Report:\n\n"
	totalTasks := len(m.tasks)
	totalSubtasks := 0
	totalCompletedSubtasks := 0

	for _, task := range m.tasks {
		totalSubtasks += task.TotalSubtasks
		totalCompletedSubtasks += task.CompletedSubtasks
		s += fmt.Sprintf("%s, Queries Completed: %d/%d\n", task.Name, task.CompletedSubtasks, task.TotalSubtasks)
	}

	s += fmt.Sprintf("\nTotal Main Tasks: %d\n", totalTasks)
	s += fmt.Sprintf("Total Subtasks Completed: %d/%d\n", totalCompletedSubtasks, totalSubtasks)
	s += "\nPress 'q' to quit."
	return s
}

func (m *model) allTasksCompleted() bool {
	for _, task := range m.tasks {
		if !task.IsCompleted {
			return false
		}
	}
	return true
}

// NEW: Function to add result tasks
func (m *model) addResultTasks() []tea.Cmd {
	var cmds []tea.Cmd
	resultTaskNames := []string{"Writing results to JSON file", "Writing results to Excel file"}

	for _, name := range resultTaskNames {
		resultTask := &MainTask{
			Name:              name,
			TotalSubtasks:     1,
			CompletedSubtasks: 0,
			Spinner:           spinner.New(),
			IsCompleted:       false,
		}
		resultTask.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		resultTask.Spinner.Spinner = spinner.Line

		m.tasks = append(m.tasks, resultTask)
		cmds = append(cmds, resultTask.Spinner.Tick)
		cmds = append(cmds, simulateResultTask(resultTask))
	}

	return cmds
}

// NEW: Function to simulate result task execution
func simulateResultTask(task *MainTask) tea.Cmd {
	return func() tea.Msg {
		// Simulate some work
		time.Sleep(time.Second * 2)
		return subtaskCompleteMsg{mainTaskName: task.Name}
	}
}

func runApp() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen())

	go func() {
		simulateTaskAddition(p)
	}()

	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func simulateTaskAddition(p *tea.Program) {
	mainTaskNames := []string{"Task A", "Task B", "Task C", "Task D", "Task E", "Task F"}

	for _, name := range mainTaskNames {
		time.Sleep(time.Second)

		mainTask := &MainTask{
			Name:              name,
			TotalSubtasks:     rand.Intn(100) + 5,
			CompletedSubtasks: 0,
			Spinner:           spinner.New(),
			IsCompleted:       false,
		}

		mainTask.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		mainTask.Spinner.Spinner = spinner.Line

		p.Send(addMainTaskMsg{mainTask: mainTask})

		go executeSubtasks(p, mainTask)
	}
}

func executeSubtasks(p *tea.Program, mainTask *MainTask) {
	for i := 0; i < mainTask.TotalSubtasks; i++ {
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
		p.Send(subtaskCompleteMsg{mainTaskName: mainTask.Name})
	}

	p.Send(mainTaskCompleteMsg{mainTaskName: mainTask.Name})
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "An application with Bubble Tea UI",
	Run: func(cmd *cobra.Command, args []string) {
		runApp()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
