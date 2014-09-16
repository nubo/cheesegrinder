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

var _ = Describe("Cheesegrinder E2E", func() {
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

		sub := Subscribe(factory, "test")
		defer closeSubscription(sub)
		go func() {
			defer GinkgoRecover()
			err := Publish(conn, "test", expectedMessage)
			Ω(err).Should(BeNil())
		}()

		msg := <-sub.Messages
		Ω(msg).Should(Equal(expectedMessage))
	})

	It("should publish a message to all subscribers", func() {
		expectedMessage := "test message"

		sub1 := Subscribe(factory, "test")
		defer closeSubscription(sub1)
		sub2 := Subscribe(factory, "test")
		defer closeSubscription(sub2)
		go func() {
			defer GinkgoRecover()
			err := Publish(conn, "test", expectedMessage)
			Ω(err).Should(BeNil())
		}()

		msg := <-sub1.Messages
		Ω(msg).Should(Equal(expectedMessage))
		msg = <-sub2.Messages
		Ω(msg).Should(Equal(expectedMessage))
	})

	AfterEach(func() {
		conn.Close()
	})
})

var _ = Describe("Cheesegrinder", func() {
	It("should close the underlying connection on close", func() {
		closed := false
		subscribed := false
		mock := RedisMock{
			SendFunc: func(cmdName string, args ...interface{}) error {
				return nil
			},
			ReceiveFunc: func() (interface{}, error) {
				if subscribed {
					select {}
				}
				subscribed = true
				return []interface{}{[]byte("subscribe"), []byte("test"), int64(1)}, nil
			},
			CloseFunc: func() error {
				closed = true
				return nil
			},
		}
		factory := func() redis.Conn {
			return mock
		}

		sub := Subscribe(factory, "test")
		close(sub.Close)
		Eventually(func() bool {
			return closed
		}).Should(BeTrue())
	})
})

func closeSubscription(s Subscription) {
	close(s.Close)
}
