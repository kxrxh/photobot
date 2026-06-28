package server

import (
	"csort.ru/coffeebot/internal/authz"
	"csort.ru/coffeebot/internal/database"
	"csort.ru/coffeebot/internal/minio"
	"csort.ru/coffeebot/internal/note"
	"csort.ru/coffeebot/internal/proposal"
	"csort.ru/coffeebot/internal/transport/http"
	"csort.ru/coffeebot/internal/weed"
	"csort.ru/coffeebot/internal/weed/analysis"
	"csort.ru/coffeebot/internal/weed/image"
	"csort.ru/coffeebot/internal/weed/stats"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Handlers struct {
	WeedHandler           *http.WeedHandler
	NotesHandler          *http.NotesHandler
	ClassificationHandler *http.ClassificationHandler
	ProposalHandler       *http.ProposalHandler
}

type Services struct {
	WeedService         *weed.Service
	WeedImageService    *image.Service
	WeedAnalysesService *analysis.Service
	WeedStatsService    *stats.Service
	NotesService        *note.NotesService
	ProposalService     *proposal.Service
	IdentityClient      authz.IdentityClient
	MinioClient         *minio.Client
}

func initializeServices(
	dbPool *pgxpool.Pool,
	minioClient *minio.Client,
	identityClient authz.IdentityClient,
	log zerolog.Logger,
) *Services {
	dbQueries := database.New(dbPool)

	weedService := weed.NewServiceWithPool(dbQueries, minioClient, dbPool, log)
	weedImageService := image.NewService(
		dbQueries,
		minioClient,
		log.With().Str("component", "services.weed_images").Logger(),
	)
	weedAnalysesService := analysis.NewService(dbQueries)
	weedStatsService := stats.NewService(dbQueries, dbPool)
	notesService := note.NewNotesService(dbQueries)
	proposalTxRunner := database.NewPoolTxRunner(dbPool, dbQueries)
	proposalService := proposal.NewService(
		dbQueries,
		proposalTxRunner,
		minioClient,
		weedService,
		weedStatsService,
		weedAnalysesService,
		weedImageService,
	)

	return &Services{
		WeedService:         weedService,
		WeedImageService:    weedImageService,
		WeedAnalysesService: weedAnalysesService,
		WeedStatsService:    weedStatsService,
		NotesService:        notesService,
		ProposalService:     proposalService,
		IdentityClient:      identityClient,
		MinioClient:         minioClient,
	}
}

func initializeHandlers(services *Services, log zerolog.Logger) *Handlers {
	return &Handlers{
		WeedHandler: http.NewWeedHandler(
			services.WeedService,
			services.WeedImageService,
			services.WeedAnalysesService,
			log.With().Str("component", "http.transport.weed").Logger(),
		),
		NotesHandler: http.NewNotesHandler(
			services.NotesService,
			log.With().Str("component", "http.transport.notes").Logger(),
		),
		ClassificationHandler: http.NewClassificationHandler(),
		ProposalHandler: http.NewProposalHandler(
			services.ProposalService,
			log.With().Str("component", "http.transport.proposals").Logger(),
		),
	}
}
