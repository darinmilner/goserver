document
  .getElementById("check-availability-btn")
  .addEventListener("click", function () {
    //notify("This is a test msg", "error");

    let html = `
            <form id="checkAvailabiltyForm" action="" method="post" novalidate class="needsValidation">
                <div class="form-row">
                    <div class="col">
                        <div class="form-row" id="reservation-dates-modal">
                            <div class="col">
                                <input disabled required class="form-control" type="text" name="start" id="start" placeholder="Arrival">    
                            </div>
                            <div class="col">
                                <input disabled required class="form-control" type="text" name="end" id="end" placeholder="Departure">
                            </div>
                        </div>
                    </div>    
                </div>
                </form>
           `;
    attention.custom({ msg: html, title: "Choose Your Dates" });
    // notifyModal("title", "<em>Hello</em>","success", "My Text for the button");
  });
