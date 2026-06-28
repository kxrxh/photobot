package routes

import "csort.ru/classification-service/internal/transport/http"

type Handlers struct {
	ClassificationHandler           *http.ClassificationHandler
	ProductHandler                  *http.ProductHandler
	UserActiveClassificationHandler *http.UserActiveClassificationHandler
	MarkupHandler                   *http.MarkupHandler
	CorrelationHandler              *http.CorrelationHandler
	ClassificationParamsHandler     *http.ClassificationParamsHandler
	OwnershipHandler                *http.OwnershipHandler
}
