package controllers

import (
	"github.com/revel/revel"
	"github.com/revel/revel/modules/jobs/app/jobs"
	"time"
)

func init() {
	// Set up the database.
	revel.OnAppStart(InitDB)

	// Start running the probes.
	revel.OnAppStart(func() {
		jobs.Every(10*time.Second, ProbeJob{})
	})

	// Before a request, make sure the right API token is set.
	revel.InterceptMethod((*GorpController).CheckToken, revel.BEFORE)

	// Before a request, we want to start a transaction.
	revel.InterceptMethod((*GorpController).Begin, revel.BEFORE)

	//	revel.InterceptMethod(Application.AddUser, revel.BEFORE)
	//	revel.InterceptMethod(Hotels.checkUser, revel.BEFORE)

	// When a request is done, we want to commit the transaction.
	revel.InterceptMethod((*GorpController).Commit, revel.AFTER)
	revel.InterceptMethod((*GorpController).Rollback, revel.FINALLY)
}
