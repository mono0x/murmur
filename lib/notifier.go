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
	urls, err := n.Sink.RecentUrls()
	if err != nil {
		return err
	}

	items, err := n.Source.Items(urls)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := n.Sink.Output(item); err != nil {
			return err
		}
	}
	return nil
}
