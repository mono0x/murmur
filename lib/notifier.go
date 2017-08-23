package murmur

import "golang.org/x/sync/errgroup"

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
	eg := errgroup.Group{}

	var items []*Item
	eg.Go(func() error {
		var err error
		items, err = n.Source.Items()
		return err
	})

	var urls []string
	eg.Go(func() error {
		var err error
		urls, err = n.Sink.RecentUrls()
		return err
	})

	if err := eg.Wait(); err != nil {
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
