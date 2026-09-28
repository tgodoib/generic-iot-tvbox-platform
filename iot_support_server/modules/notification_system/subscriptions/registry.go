package subscriptions

import (
	"notification_system/platforms"
	"notification_system/record"
)

const ALL = "*"

type Registry struct {
	subs map[record.AlertType]map[string]platforms.NotificationPlatform
}

func NewRegistry() *Registry {
	return &Registry{
		subs: make(map[record.AlertType]map[string]platforms.NotificationPlatform),
	}
}

func (r *Registry) Register(plat platforms.NotificationPlatform, types []record.AlertType, devices []string) {
	for _, t := range types {
		if _, ok := r.subs[t]; !ok {
			r.subs[t] = make(map[string]platforms.NotificationPlatform)
		}
		for _, d := range devices {
			r.subs[t][d] = plat
		}
	}
}

func (r *Registry) Alert(rec record.AlertRecord) []error {
	errs := make([]error, 0)

	typesToCheck := []record.AlertType{rec.Type, ALL}
	devicesToCheck := []string{rec.DeviceID, ALL}

	seen := make(map[platforms.NotificationPlatform]bool)

	for _, t := range typesToCheck {

		if deviceSubs, ok := r.subs[t]; ok {
			for _, d := range devicesToCheck {

				if plat, ok := deviceSubs[d]; ok {
					if seen[plat] {
						continue
					}
					seen[plat] = true

					if err := plat.Send(rec.Title, rec.Body, rec.DeviceID); err != nil {
						errs = append(errs, err)
					}
				}

			}
		}

	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
