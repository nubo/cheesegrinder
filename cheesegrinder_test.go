package cheesegrinder

import (
	"os"

	"github.com/garyburd/redigo/redis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestPubSub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cheesegrinder")
}

var _ = Describe("Cheesegrinder", func() {
	if os.Getenv("TEST_REDIS") == "" {
		GinkgoT().Skipf("[!!] TEST_REDIS environment variable not set. Skipping e2e tests\n")
		return
	}

	var conn redis.Conn
	factory := func() redis.Conn {
		conn, err := redis.Dial("tcp", os.Getenv("TEST_REDIS"))
		if err != nil {
			panic(err)
		}
		return conn
	}

	BeforeEach(func() {
		defer GinkgoRecover()
		conn = factory()
	})

	It("should publish a message to a subscriber", func() {
		expectedMessage := "test message"

		subscribed := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			<-subscribed
			err := Publish(conn, "test", expectedMessage)
			Ω(err).Should(BeNil())
		}()

		sub := Subscribe(factory, "test")
		close(subscribed)
		msg := <-sub.Messages
		Ω(msg).Should(Equal(expectedMessage))
	})

	AfterEach(func() {
		conn.Close()
	})
})
