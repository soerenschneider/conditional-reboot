package checkers

import (
	"errors"
	"fmt"
	"time"
)

func UseTLS(certFile, keyFile string) KafkaOpts {
	return func(c *KafkaChecker) error {
		c.certFile = certFile
		c.keyFile = keyFile
		return nil
	}
}

func WithGraceTime(gracetime time.Duration) KafkaOpts {
	return func(c *KafkaChecker) error {
		if gracetime < defaultGraceTime {
			return fmt.Errorf("gracetime must not be < %v", defaultGraceTime)
		}

		c.graceTime = gracetime
		return nil
	}
}

func AcceptedKeys(keys []string) KafkaOpts {
	return func(c *KafkaChecker) error {
		if len(keys) == 0 {
			return errors.New("empty slice provided as kafka keys")
		}

		acceptedKeys := map[string]bool{}
		for _, key := range keys {
			acceptedKeys[key] = true
		}

		c.acceptedKeys = acceptedKeys
		return nil
	}
}
