package models

// Job represents a download task
type Job struct {
	URL string
	ID  int
}

// Todo represents the data we'll insert into Postgres
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}
