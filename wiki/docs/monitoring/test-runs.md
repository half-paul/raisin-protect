# Test Runs

Tests check whether your controls are working correctly. You can run them manually or let them run on a schedule.

## Running All Tests

1. Go to **Monitoring > Test Runs** in the sidebar
2. Click **"Run All Tests"**
3. Confirm in the dialog by clicking **"Run Tests"**
4. A new test run appears in the list with status **Pending**, then **Running**
5. When complete, the status changes to **Completed** with pass/fail counts

## Viewing Test Run History

The Test Runs page shows all previous runs with:

- **Run number** and status badge
- **Trigger type** — Manual, Scheduled, On Change, or Webhook
- **Results** — Pass, Fail, and Error counts
- **Duration** — How long the run took
- **Triggered By** — Who or what started it

Use the **Status** and **Trigger** filters to narrow down the list.

!!! note "Permissions"
    CISO, Compliance Manager, Security Engineer, and DevOps Engineer can run tests.

## Viewing Test Results

1. Click the link icon on any test run row, or navigate to a test run's detail page
2. See summary cards: Total Tests, Passed (green), Failed (red), Errors (orange), Skipped/Warnings
3. Use the filter buttons to show only failed, errored, passing, or all results
4. Each result shows:
   - **Status** — Pass, Fail, Error, Skip, or Warning
   - **Severity** — How critical the test is
   - **Test ID and title** — Which test ran
   - **Control** — Which control this test checks (clickable)
   - **Message** — Brief result summary
   - **Alert** — If a failure generated an alert (clickable link to the alert)
5. Click the detail icon on any result to see the full output log, error messages, and JSON details
