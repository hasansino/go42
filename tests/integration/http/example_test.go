// nolint
package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hasansino/go42/tests/integration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type Fruit struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateFruitRequest struct {
	Name string `json:"name"`
}

type UpdateFruitRequest struct {
	Name string `json:"name"`
}

var _ = Describe("Fruits API Integration Tests", func() {
	var client *http.Client

	BeforeEach(func() {
		client = &http.Client{Timeout: 5 * time.Second}
	})

	Describe("GET /fruits", func() {
		It("should return a list of fruits", func() {
			resp, err := client.Get(integration.HTTPServerAddress() + "/fruits?limit=5&offset=0")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var fruits []Fruit
			err = json.NewDecoder(resp.Body).Decode(&fruits)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(fruits)).To(BeNumerically("<=", 5))
		})
	})

	Describe("POST /fruits", func() {
		It("should create a new fruit and cleanup after itself", func() {
			name := integration.GenerateRandomName("mango")
			reqBody := CreateFruitRequest{Name: name}
			bodyBytes, err := json.Marshal(reqBody)
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(
				integration.HTTPServerAddress()+"/fruits",
				"application/json",
				bytes.NewReader(bodyBytes),
			)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var createdFruit Fruit
			err = json.NewDecoder(resp.Body).Decode(&createdFruit)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdFruit.Name).To(Equal(name))
			Expect(createdFruit.ID).To(BeNumerically(">", 0))

			// Cleanup
			req, err := http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), createdFruit.ID),
				nil,
			)
			Expect(err).ToNot(HaveOccurred())
			delResp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer delResp.Body.Close()
			Expect(delResp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("GET /fruits/{id}", func() {
		It("should create, get by ID and cleanup", func() {
			name := integration.GenerateRandomName("apple")
			// CreateFruit fruit
			reqBody := CreateFruitRequest{Name: name}
			bodyBytes, err := json.Marshal(reqBody)
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(
				integration.HTTPServerAddress()+"/fruits",
				"application/json",
				bytes.NewReader(bodyBytes),
			)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var fruit Fruit
			err = json.NewDecoder(resp.Body).Decode(&fruit)
			Expect(err).ToNot(HaveOccurred())

			// Get by ID
			getResp, err := client.Get(fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), fruit.ID))
			Expect(err).ToNot(HaveOccurred())
			defer getResp.Body.Close()
			Expect(getResp.StatusCode).To(Equal(http.StatusOK))

			var fetchedFruit Fruit
			err = json.NewDecoder(getResp.Body).Decode(&fetchedFruit)
			Expect(err).ToNot(HaveOccurred())
			Expect(fetchedFruit.ID).To(Equal(fruit.ID))
			Expect(fetchedFruit.Name).To(Equal(fruit.Name))

			// Cleanup
			req, err := http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), fruit.ID),
				nil,
			)
			Expect(err).ToNot(HaveOccurred())
			delResp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer delResp.Body.Close()
			Expect(delResp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should return 404 for non-existing fruit", func() {
			resp, err := client.Get(fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), 999999))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("PUT /fruits/{id}", func() {
		It("should create, update and cleanup a fruit", func() {
			name := integration.GenerateRandomName("banana")
			// CreateFruit fruit
			reqBody := CreateFruitRequest{Name: name}
			bodyBytes, err := json.Marshal(reqBody)
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(
				integration.HTTPServerAddress()+"/fruits",
				"application/json",
				bytes.NewReader(bodyBytes),
			)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var fruit Fruit
			err = json.NewDecoder(resp.Body).Decode(&fruit)
			Expect(err).ToNot(HaveOccurred())

			// UpdateFruit fruit
			updatedName := integration.GenerateRandomName("kek")
			updateReq := UpdateFruitRequest{Name: updatedName}
			updateBytes, err := json.Marshal(updateReq)
			Expect(err).ToNot(HaveOccurred())

			req, err := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), fruit.ID),
				bytes.NewReader(updateBytes),
			)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")

			updateResp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer updateResp.Body.Close()
			Expect(updateResp.StatusCode).To(Equal(http.StatusOK))

			var updatedFruit Fruit
			err = json.NewDecoder(updateResp.Body).Decode(&updatedFruit)
			Expect(err).ToNot(HaveOccurred())
			Expect(updatedFruit.Name).To(Equal(updatedName))
			Expect(updatedFruit.ID).To(Equal(fruit.ID))

			// Cleanup
			delReq, err := http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), fruit.ID),
				nil,
			)
			Expect(err).ToNot(HaveOccurred())
			delResp, err := client.Do(delReq)
			Expect(err).ToNot(HaveOccurred())
			defer delResp.Body.Close()
			Expect(delResp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should return 404 when updating non-existing fruit", func() {
			updateReq := UpdateFruitRequest{Name: integration.GenerateRandomName("nonexistent")}
			bodyBytes, err := json.Marshal(updateReq)
			Expect(err).ToNot(HaveOccurred())

			req, err := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), 999999),
				bytes.NewReader(bodyBytes),
			)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("DELETE /fruits/{id}", func() {
		It("should create and delete a fruit", func() {
			name := integration.GenerateRandomName("peach")
			// CreateFruit fruit
			reqBody := CreateFruitRequest{Name: name}
			bodyBytes, err := json.Marshal(reqBody)
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(
				integration.HTTPServerAddress()+"/fruits",
				"application/json",
				bytes.NewReader(bodyBytes),
			)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var fruit Fruit
			err = json.NewDecoder(resp.Body).Decode(&fruit)
			Expect(err).ToNot(HaveOccurred())

			// DeleteFruit fruit
			req, err := http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), fruit.ID),
				nil,
			)
			Expect(err).ToNot(HaveOccurred())

			delResp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer delResp.Body.Close()
			Expect(delResp.StatusCode).To(Equal(http.StatusOK))

			// Verify deletion
			getResp, err := client.Get(fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), fruit.ID))
			Expect(err).ToNot(HaveOccurred())
			defer getResp.Body.Close()
			Expect(getResp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return 404 when deleting non-existing fruit", func() {
			req, err := http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("%s/fruits/%d", integration.HTTPServerAddress(), 999999),
				nil,
			)
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fruits API Integration Suite")
}
