package cheesegrinder

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

// Subscription is a subscription to a topic. All messages will
// be pushed to the Messages channel. Close the close-cha
type Subscription struct {
	Messages <-chan string
	c        chan struct{}
}

// Close unsubscribes from the topic and closes the connection.
func (s *Subscription) Close() {
	close(s.c)
}

// ConnectionFactory is a function which creates a new redis.Conn.
// Since a connection can't be used once it is in subscription mode,
// this pattern has been introduced to create a new connection for
// every subscription and give an opportunity to do authentication.
type ConnectionFactory func() redis.Conn

// Subscribe creates a new subscription to the given topic.
func Subscribe(f ConnectionFactory, topic string) Subscription {
	closing := make(chan struct{})
	msgs := make(chan string)
	subscribed := make(chan struct{})
	rawmsgs := make(chan interface{})

	go func() {
		psc := redis.PubSubConn{f()}
		defer psc.Close()
		psc.Subscribe(topic)
		defer psc.Unsubscribe(topic)

		go func() {
			defer close(rawmsgs)
			for {
				rawmsg := psc.Receive()
				rawmsgs <- rawmsg
			}
		}()
		<-closing
	}()

	go func() {
		for {
			var plainMsg []byte
			select {
			case rawmsg := <-rawmsgs:
				switch msg := rawmsg.(type) {
				case redis.Subscription:
					close(subscribed)
					continue
				case redis.Message:
					plainMsg = msg.Data
				case error:
					log.Printf("Error: %s", msg)
					return
				default:
					log.Printf("Unknown type %#v", rawmsg)
					continue
				}
			case <-closing:
				return
			}

			select {
			case msgs <- string(plainMsg):
			case <-closing:
				return
			}
		}
	}()

	<-subscribed
	return Subscription{
		Messages: msgs,
		c:        closing,
	}
}

// Publish publishes a single message on the given topic.
func Publish(c redis.Conn, topic string, msg string) error {
	_, err := c.Do("PUBLISH", topic, msg)
	return err
}
