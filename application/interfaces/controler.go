package interfaces

import (
	"TML_TBot/domain/models"
)

type UseCase interface {
	Run() ([]models.TGMessage, error)
}
