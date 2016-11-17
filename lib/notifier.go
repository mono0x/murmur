package murmur

type Notifier struct {
	Source Source
	Sink   Sink
}

func NewNotifier(source Source, sink Sink) *Notifier {
	return &Notifier{
		Source: source,
		Sink:   sink,
	}
}

func (n *Notifier) Notify() error {
	items, err := n.Source.Items()
	if err != nil {
		return err
	}

	urls, err := n.Sink.RecentUrls()
	if err != nil {
		return err
	}

	duplications := make(map[string]struct{}, len(urls)+len(items))
	for _, url := range urls {
		duplications[url] = struct{}{}
	}

	for _, item := range items {
		if _, ok := duplications[item.Url]; ok {
			continue
		}
		duplications[item.Url] = struct{}{}

		if err := n.Sink.Output(item); err != nil {
			return err
		}
	}
	return nil
}
