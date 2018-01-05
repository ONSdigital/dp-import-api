package event

import (
	"github.com/ONSdigital/dp-import-api/job"
	"github.com/ONSdigital/dp-import/events"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

// JobService provide business logic for job related operations.
type JobService interface {
	UpdateInstanceTaskState(jobID, instanceID, taskID, newState string) error
}

func StartConsumer(observationsImportedConsumer *kafka.ConsumerGroup, jobService JobService) chan bool {
	stopHandlingConsumerChan := make(chan bool)
	go func() {
		for {
			select {
			case observationsImportedMessage := <-observationsImportedConsumer.Incoming():
				logData := log.Data{"kafka_offset": observationsImportedMessage.Offset()}
				if err := Consume(observationsImportedMessage, jobService); err != nil {
					observationsImportedConsumer.Release()
					log.ErrorC("event processed - not committing message", err, logData)
				} else {
					observationsImportedConsumer.CommitAndRelease(observationsImportedMessage)
					log.Debug("event processed - message committed", logData)
				}
			case <-stopHandlingConsumerChan:
				// channel closed, so stop processing
				return
			}
		}
	}()
	return stopHandlingConsumerChan
}

// Consume takes a message, unmarshalls and sends it to the jobService
func Consume(message kafka.Message, jobService JobService) error {
	var event events.ObservationImportComplete
	logData := log.Data{"kafka_offset": message.Offset()}
	err := events.ObservationImportCompleteSchema.Unmarshal(message.GetData(), &event)
	if err != nil {
		log.ErrorC("Consume failed to unmarshal event", err, logData)
		return err
	}

	logData["event"] = event
	log.Debug("event received", logData)

	err = jobService.UpdateInstanceTaskState(
		event.JobID,
		event.InstanceID,
		job.ImportTaskIDImportObservations,
		job.ImportTaskStateComplete)

	if err != nil {
		// TODO - set task / job state to failed?
		log.ErrorC("Consume failed to handle event", err, logData)
		return err
	}

	return nil
}
