package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/drive"
	"github.com/gofiber/fiber/v2"
)

type DetaDrive struct {
	drive *drive.Drive
}

func main() {
	key := os.Getenv("key")
	if key == "" {
		key = ""
	}

	d, err := deta.New(deta.WithProjectKey(key))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create new Deta instance")
		os.Exit(1)
	}

	ds, err := drive.New(d, "storage")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create new Drive instance: %v\n", err)
	}

	dd := &DetaDrive{
		drive: ds,
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("index.html")
	})

	app.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		fileContent, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		defer fileContent.Close()

		err = dd.Put(fileContent, file.Filename, file.Header.Get("Content-Type"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.SendString("File uploaded successfully!")
	})

	app.Listen(":3000")
}

func (dd *DetaDrive) Put(file io.Reader, filename string, contentType string) error {
	_, err := dd.drive.Put(&drive.PutInput{
		Name:        filename,
		Body:        bufio.NewReader(file),
		ContentType: contentType,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to put file: %v\n", err)
		return err
	}

	fmt.Printf("successfully put file %s", filename)
	return nil
}
