package cheesegrinder

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

type Subscription struct {
	Messages <-chan string
	Close    chan struct{}
}

// ConnectionFactory is a function which creates a new redis.Conn.
// Since a connection can't be used once it is in subscription mode,
// this pattern has been introduced to create a new connection for
// every subscription and give an opportunity to do authentication.
type ConnectionFactory func() redis.Conn

func Subscribe(f ConnectionFactory, topic string) Subscription {
	running := true

	closing := make(chan struct{})
	go func() {
		<-closing
		running = false
	}()

	msgs := make(chan string)
	subscribed := make(chan struct{})
	go func() {
		psc := redis.PubSubConn{f()}
		defer psc.Close()
		psc.Subscribe(topic)
		defer psc.Unsubscribe(topic)
		for running {
			rawmsg := psc.Receive()
			switch msg := rawmsg.(type) {
			case redis.Subscription:
				close(subscribed)
			case redis.Message:
				select {
				case msgs <- string(msg.Data):
				case <-closing:
				}
			case error:
				log.Printf("Error: %s", msg)
				return
			default:
				log.Printf("Unknown type %#v", rawmsg)
			}
		}
	}()

	<-subscribed
	return Subscription{
		Messages: msgs,
		Close:    closing,
	}
}

func Publish(c redis.Conn, topic string, msg string) error {
	_, err := c.Do("PUBLISH", topic, msg)
	return err
}
