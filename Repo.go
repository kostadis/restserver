package main 

import "fmt" 

var currentID int

var todos Todos

func init() { 
    RepoCreateTodo(Todo{Name: "Write presentation"})
    RepoCreateTodo(Todo{Name: "Host meetup"})
}

func RepoFindTodo(id int) Todo {
    for _, t := range todos { 
        if t.ID == id { 
            return t
        }
    }
    // return empty todo if not found 
    return Todo {}
}

func RepoCreateTodo(t Todo) Todo { 
    currentID += 1
    t.ID = currentID
    todos = append(todos, t)
    return t
}

func RepoDestroyTodo(id int) error { 
    for i, t := range todos { 
        if t.ID == id { 
            todos = append(todos[:i], todos[i+1:]...)
            return nil
        }
    }
    return fmt.Errorf("Could not find Todo with id of %d to delete", id)
}