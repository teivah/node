package main

import (
	"context"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"log"
	"path"
)

func main() {

	e2eProject, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: []string{path.Join("e2e", "docker-compose.yml")},
			ProjectName:  "e2e-test",
		},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer e2eProject.Down(context.Background(), options.Down{})

	err = e2eProject.Up(context.Background(), options.Up{}, "db")
	if err != nil {
		log.Println(err)
	}
	log.Println("Database started")

	err = e2eProject.Build(context.Background(), options.Build{}, "node", "client")
	if err != nil {
		log.Println(err)
	}
	log.Println("Images built")

	exitCode, err := e2eProject.Run(
		context.Background(),
		"discovery",
		[]string{
			"--entrypoint",
			"bin/db-upgrade",
		},
		options.Run{},
	)
	if err != nil {
		log.Println(err)
	}
	log.Println("Migration completed. Exit code: ", exitCode)

	err = e2eProject.Up(context.Background(), options.Up{})
	if err != nil {
		log.Println(err)
	}

	log.Println("All services are up - ready for something")

}
