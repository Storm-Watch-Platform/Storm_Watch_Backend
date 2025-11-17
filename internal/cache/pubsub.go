package cache

func Subscribe(channel string, handler func(msg string)) {
	sub := Rdb.Subscribe(Ctx, channel)
	ch := sub.Channel()

	go func() {
		for m := range ch {
			handler(m.Payload)
		}
	}()
}

func Publish(channel, msg string) error {
	return Rdb.Publish(Ctx, channel, msg).Err()
}
