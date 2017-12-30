package event

import (
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import/events"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

// Consumer consumes event messages.
type ObservationsImportedConsumer struct {
	*kafka.AsyncConsumer
}

func NewObservationsImportedConsumer() *ObservationsImportedConsumer {
	return &ObservationsImportedConsumer{
		AsyncConsumer: kafka.NewAsyncConsumer(),
	}
}

// JobService provide business logic for job related operations.
type JobService interface {
	UpdateInstanceTaskState(jobID, instanceID, taskID, newState string) error
}

func (consumer *ObservationsImportedConsumer) Consume(messageConsumer kafka.MessageConsumer, jobService JobService) {

	handlerFunc := func(message kafka.Message) {

		var event events.ObservationImportComplete
		err := events.ObservationImportCompleteSchema.Unmarshal(message.GetData(), &event)
		if err != nil {
			log.Error(err, log.Data{"message": "failed to unmarshal event"})
			return
		}

		log.Debug("event received", log.Data{"event": event})

		err = jobService.UpdateInstanceTaskState(
			event.JobID,
			event.InstanceID,
			job.ImportTaskIDImportObservations,
			job.ImportTaskStateComplete)

		if err != nil {
			// todo - set task / job state to failed?
			log.Error(err, log.Data{"message": "failed to handle event"})
		}

		log.Debug("event processed - committing message", log.Data{"event": event})
		message.Commit()
		log.Debug("message committed", log.Data{"event": event})
	}

	consumer.AsyncConsumer.Consume(messageConsumer, handlerFunc)
}
