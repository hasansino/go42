// nolint
package grpc

import (
	"context"
	"testing"

	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/hasansino/go42/internal/example/provider/grpc"
	"github.com/hasansino/go42/tests/integration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fruits gRPC Integration Tests", func() {
	var (
		conn   *grpclib.ClientConn
		client pb.ExampleServiceClient
		ctx    context.Context
	)

	BeforeEach(func() {
		var err error
		conn, err = grpclib.NewClient(
			integration.GRPCServerAddress(),
			grpclib.WithTransportCredentials(insecure.NewCredentials()),
		)
		Expect(err).NotTo(HaveOccurred())
		client = pb.NewExampleServiceClient(conn)
		ctx = context.Background()
	})

	AfterEach(func() {
		if conn != nil {
			conn.Close()
		}
	})

	Describe("ListFruits", func() {
		It("should return a list of fruits", func() {
			req := &pb.ListFruitsRequest{
				Limit:  5,
				Offset: 0,
			}

			resp, err := client.ListFruits(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(len(resp.Fruits)).To(BeNumerically("<=", 5))
		})
	})

	Describe("CreateFruit", func() {
		It("should create a new fruit and cleanup after itself", func() {
			name := integration.GenerateRandomString("mango")
			req := &pb.CreateFruitRequest{
				Name: name,
			}

			// Create fruit
			createResp, err := client.CreateFruit(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResp).NotTo(BeNil())
			Expect(createResp.Fruit).NotTo(BeNil())
			Expect(createResp.Fruit.Name).To(Equal(name))
			Expect(createResp.Fruit.Id).To(BeNumerically(">", 0))

			// Cleanup
			delReq := &pb.DeleteFruitRequest{
				Id: createResp.Fruit.Id,
			}

			delResp, err := client.DeleteFruit(ctx, delReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(delResp.Success).To(BeTrue())
		})
	})

	Describe("GetFruit", func() {
		It("should create, get by ID and cleanup", func() {
			// Create fruit first
			name := integration.GenerateRandomString("apple")
			createReq := &pb.CreateFruitRequest{
				Name: name,
			}

			createResp, err := client.CreateFruit(ctx, createReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResp.Fruit).NotTo(BeNil())
			fruitID := createResp.Fruit.Id

			// Get by ID
			getReq := &pb.GetFruitRequest{
				Id: fruitID,
			}

			fruit, err := client.GetFruit(ctx, getReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(fruit).NotTo(BeNil())
			Expect(fruit.Id).To(Equal(fruitID))
			Expect(fruit.Name).To(Equal(name))

			// Cleanup
			delReq := &pb.DeleteFruitRequest{
				Id: fruitID,
			}

			delResp, err := client.DeleteFruit(ctx, delReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(delResp.Success).To(BeTrue())
		})

		It("should return NotFound error for non-existing fruit", func() {
			getReq := &pb.GetFruitRequest{
				Id: 999999,
			}

			_, err := client.GetFruit(ctx, getReq)
			Expect(err).To(HaveOccurred())

			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})
	})

	Describe("UpdateFruit", func() {
		It("should create, update and cleanup a fruit", func() {
			// Create fruit first
			name := integration.GenerateRandomString("banana")
			createReq := &pb.CreateFruitRequest{
				Name: name,
			}

			createResp, err := client.CreateFruit(ctx, createReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResp.Fruit).NotTo(BeNil())
			fruitID := createResp.Fruit.Id

			// Update fruit
			updatedName := integration.GenerateRandomString("updated-banana")
			updateReq := &pb.UpdateFruitRequest{
				Id:   fruitID,
				Name: updatedName,
			}

			updateResp, err := client.UpdateFruit(ctx, updateReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(updateResp).NotTo(BeNil())
			Expect(updateResp.Fruit).NotTo(BeNil())
			Expect(updateResp.Fruit.Id).To(Equal(fruitID))
			Expect(updateResp.Fruit.Name).To(Equal(updatedName))

			// Cleanup
			delReq := &pb.DeleteFruitRequest{
				Id: fruitID,
			}

			delResp, err := client.DeleteFruit(ctx, delReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(delResp.Success).To(BeTrue())
		})

		It("should return NotFound when updating non-existing fruit", func() {
			updateReq := &pb.UpdateFruitRequest{
				Id:   999999,
				Name: integration.GenerateRandomString("nonexistent"),
			}

			_, err := client.UpdateFruit(ctx, updateReq)
			Expect(err).To(HaveOccurred())

			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})
	})

	Describe("DeleteFruit", func() {
		It("should create and delete a fruit", func() {
			// Create fruit first
			name := integration.GenerateRandomString("peach")
			createReq := &pb.CreateFruitRequest{
				Name: name,
			}

			createResp, err := client.CreateFruit(ctx, createReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResp.Fruit).NotTo(BeNil())
			fruitID := createResp.Fruit.Id

			// Delete fruit
			delReq := &pb.DeleteFruitRequest{
				Id: fruitID,
			}

			delResp, err := client.DeleteFruit(ctx, delReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(delResp.Success).To(BeTrue())

			// Verify deletion by trying to get it
			getReq := &pb.GetFruitRequest{
				Id: fruitID,
			}

			_, err = client.GetFruit(ctx, getReq)
			Expect(err).To(HaveOccurred())

			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})

		It("should return NotFound when deleting non-existing fruit", func() {
			delReq := &pb.DeleteFruitRequest{
				Id: 999999,
			}

			_, err := client.DeleteFruit(ctx, delReq)
			Expect(err).To(HaveOccurred())

			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})
	})
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fruits gRPC Integration Suite")
}
