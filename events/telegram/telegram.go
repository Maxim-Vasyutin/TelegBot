package telegram

import (
	"errors"

	"main.go/clients/telegram"
	"main.go/events"
	"main.go/lib/e"
	"main.go/storage"
)

// апдейты - понятие телеграмма, в другом мессанджере может его и не быть
// а ивент - более общая сущность, в неё можем преобразовывать всё что получаем от других м.
type Processor struct {
	tg      *telegram.Client
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID   int
	Username string
}

var (
	ErrUnknowEventType = errors.New("unknow event type")
	ErrUnknowMetaType  = errors.New("unknow meta type")
)

func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	//тут получаем апдейты
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}
	//тут если список апдейтов = 0, то заканчиваем работу функции
	if len(updates) == 0 {
		return nil, nil
	}
	//тут готовим переменную для результата, аллоцируем её
	res := make([]events.Event, 0, len(updates))
	//тут перебираем все апдейты и преобразуем в тип event
	for _, u := range updates {
		res = append(res, event(u))
	}
	//тут обновляем параметр offset, чтобы в следующий раз получить пачку изменений
	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	default:
		return e.Wrap("can't process message", ErrUnknowEventType)
	}
}

func (p *Processor) processMessage(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}
	
	if err := p.doCmd(event.Text, meta.ChatID, meta.Username); err != nil {
		return e.Wrap("can't process message", err)
	}
	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknowMetaType)
	}

	return res, nil
}

func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)

	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}

	return res
}

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknow
	}
	return events.Message

}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}
	return upd.Message.Text
}
