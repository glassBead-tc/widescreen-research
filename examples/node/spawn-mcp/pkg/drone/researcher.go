package drone

import "log"

type ResearcherDrone struct{}

func NewResearcherDrone() (*ResearcherDrone, error) {
	return &ResearcherDrone{}, nil
}

func (d *ResearcherDrone) Serve() error {
	log.Println("Drone running...")
	select {}
}

func (d *ResearcherDrone) Close() error {
	return nil
}