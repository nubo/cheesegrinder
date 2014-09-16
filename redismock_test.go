package cheesegrinder

type RedisMock struct {
	CloseFunc   func() error
	ErrorFunc   func() error
	DoFunc      func(commandName string, args ...interface{}) (reply interface{}, err error)
	SendFunc    func(commandName string, args ...interface{}) error
	FlushFunc   func() error
	ReceiveFunc func() (reply interface{}, err error)
}

func (rm RedisMock) Close() error {
	if rm.CloseFunc == nil {
		return nil
	}
	return rm.CloseFunc()
}

func (rm RedisMock) Err() error {
	if rm.ErrorFunc == nil {
		return nil
	}
	return rm.ErrorFunc()
}

func (rm RedisMock) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if rm.DoFunc == nil {
		return nil, nil
	}
	return rm.DoFunc(commandName, args...)
}

func (rm RedisMock) Send(commandName string, args ...interface{}) error {
	if rm.SendFunc == nil {
		return nil
	}
	return rm.SendFunc(commandName, args...)
}

func (rm RedisMock) Flush() error {
	if rm.FlushFunc == nil {
		return nil
	}
	return rm.FlushFunc()
}

func (rm RedisMock) Receive() (reply interface{}, err error) {
	if rm.ReceiveFunc == nil {
		return nil, nil
	}
	return rm.ReceiveFunc()
}
