package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestTheFirstSuccessfulScanStartsTheFirewallEvaluation(testRun *testing.T) {
	clock := &testClock{now: testRefreshTime}
	evaluator := &fakeFirewall{report: sampleFirewallReport()}
	model, command := firewallTestModel(testRun, evaluator, clock)
	if command == nil || !model.firewallRefreshing {
		testRun.Fatal("the first scan result must start the firewall evaluation")
	}
	message, isFirewall := command().(FirewallResultMessage)
	if !isFirewall || message.Err != nil || evaluator.calls != 1 || !message.At.Equal(clock.now) {
		testRun.Fatalf("message=%+v calls=%d", message, evaluator.calls)
	}
	if len(evaluator.ports) != len(sampleReportWithQuietPort().Ports) {
		testRun.Fatalf("the current host ports must be evaluated, got %d", len(evaluator.ports))
	}
}

func TestNothingIsEvaluatedBeforeThereIsAScanOrWhileOneRuns(testRun *testing.T) {
	clock := &testClock{now: testRefreshTime}
	model := New(Options{Scanner: fakeScanner{}, Firewall: &fakeFirewall{}, Interval: testInterval, Version: testVersion, Clock: clock.read})
	ticked, _ := updateWith(testRun, model, tickMessage{})
	if ticked.firewallRefreshing {
		testRun.Fatal("no report yet: nothing to evaluate")
	}
	failed, command := updateWith(testRun, model, RefreshResultMessage{Err: errDockerDown, At: clock.now})
	if failed.firewallRefreshing || command != nil {
		testRun.Fatal("a failed scan must not start the firewall evaluation")
	}
	started, _ := firewallTestModel(testRun, &fakeFirewall{}, clock)
	_, second := updateWith(testRun, started, RefreshResultMessage{Report: sampleReportWithQuietPort(), At: clock.now})
	if second != nil {
		testRun.Fatal("an evaluation already in flight must not be started twice")
	}
}

func TestTheFirewallIsThrottledIndependentlyFromTheScanInterval(testRun *testing.T) {
	clock := &testClock{now: testRefreshTime}
	model, _ := firewallTestModel(testRun, &fakeFirewall{report: sampleFirewallReport()}, clock)
	model = applyMessage(testRun, model, FirewallResultMessage{Report: sampleFirewallReport(), At: clock.now})
	clock.advance(firewallRefreshInterval - time.Second)
	model = applyMessage(testRun, model, tickMessage{})
	if model.firewallRefreshing {
		testRun.Fatal("a tick inside the throttle window must not read the firewall")
	}
	model, _ = updateWith(testRun, model, RefreshResultMessage{Report: sampleReportWithQuietPort(), At: clock.now})
	if model.firewallRefreshing {
		testRun.Fatal("a scan with the same ports inside the window must not read the firewall")
	}
	clock.advance(2 * time.Second)
	model = applyMessage(testRun, model, tickMessage{})
	if !model.firewallRefreshing {
		testRun.Fatal("a tick after the throttle window must read the firewall again")
	}
}

func TestRefreshKeyReevaluatesTheFirewallInsideTheThrottleWindow(testRun *testing.T) {
	clock := &testClock{now: testRefreshTime}
	model, _ := firewallTestModel(testRun, &fakeFirewall{report: sampleFirewallReport()}, clock)
	model = applyMessage(testRun, model, FirewallResultMessage{Report: sampleFirewallReport(), At: clock.now})
	clock.advance(time.Second)
	model = typeKeys(testRun, model, "r")
	if !model.firewallRefreshing {
		testRun.Fatal("r must re-evaluate the firewall")
	}
}

func TestAChangeOfThePortSetReevaluatesTheFirewallAtOnce(testRun *testing.T) {
	clock := &testClock{now: testRefreshTime}
	model, _ := firewallTestModel(testRun, &fakeFirewall{report: sampleFirewallReport()}, clock)
	model = applyMessage(testRun, model, FirewallResultMessage{Report: sampleFirewallReport(), At: clock.now})
	clock.advance(time.Second)
	model, command := updateWith(testRun, model, RefreshResultMessage{Report: reportWith(sampleReportWithQuietPort(), freshPortNumber), At: clock.now})
	if !model.firewallRefreshing || command == nil {
		testRun.Fatal("a new listening port must trigger an evaluation without waiting")
	}
}

func TestAnExternalPortDoesNotChangeTheFirewallPortSet(testRun *testing.T) {
	external := domain.Port{Number: 8443, Protocol: domain.TCP, BindAddress: "203.0.113.10", Source: domain.SourceExternal}
	host := sampleReport().Ports
	if firewallPortSignature(host) != firewallPortSignature(append(append([]domain.Port{}, host...), external)) {
		testRun.Fatal("external ports are not evaluated, so they must not change the signature")
	}
	reversed := append([]domain.Port{}, host...)
	reversed[0], reversed[1] = reversed[1], reversed[0]
	if firewallPortSignature(host) != firewallPortSignature(reversed) {
		testRun.Fatal("the order of the ports must not matter")
	}
}

func TestAFailedEvaluationKeepsThePreviousDataAndWarnsOnTheFirewallViews(testRun *testing.T) {
	model := evaluatedModel(testRun, sampleFirewallReport())
	model = applyMessage(testRun, model, FirewallResultMessage{Err: errFirewallDown, At: testRefreshTime.Add(time.Minute)})
	if !model.hasFirewall || len(model.firewallReport.Findings) != len(sampleFirewallReport().Findings) || model.firewallRefreshing {
		testRun.Fatal("a failed evaluation must keep the previous data")
	}
	if !model.firewallAt.Equal(testRefreshTime) {
		testRun.Fatalf("the evaluation time must stay at the last good one, got %v", model.firewallAt)
	}
	for _, inner := range []InnerTab{0, 1} {
		if rendered := showView(testRun, model, moduleView{ModuleFirewall, inner}).View(); !strings.Contains(rendered, "firewall refresh failed") {
			testRun.Errorf("inner %d lacks the warning:\n%s", inner, rendered)
		}
	}
	if strings.Contains(showModule(testRun, model, ModulePorts).View(), "firewall refresh failed") {
		testRun.Error("the warning only belongs to the Firewall views")
	}
	recovered := applyMessage(testRun, model, FirewallResultMessage{Report: sampleFirewallReport(), At: testRefreshTime.Add(2 * time.Minute)})
	if strings.Contains(showModule(testRun, recovered, ModuleFirewall).View(), "firewall refresh failed") {
		testRun.Error("a good evaluation must clear the warning")
	}
}

func TestAFailureBeforeAnyDataSaysSo(testRun *testing.T) {
	clock := &testClock{now: testRefreshTime}
	model, _ := firewallTestModel(testRun, &fakeFirewall{}, clock)
	model = applyMessage(testRun, model, FirewallResultMessage{Err: errFirewallDown, At: clock.now})
	rendered := showView(testRun, model, moduleView{ModuleFirewall, 1}).View()
	for _, want := range []string{"could not be evaluated yet", "firewall refresh failed"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("view lacks %q:\n%s", want, rendered)
		}
	}
}

func TestAnUnavailableFirewallIsANormalStateShownWithItsNotes(testRun *testing.T) {
	model := evaluatedModel(testRun, unavailableFirewallReport())
	status := showModule(testRun, model, ModuleFirewall).View()
	for _, want := range []string{"Available: no", "needs root to read firewall rules", "unknown"} {
		if !strings.Contains(status, want) {
			testRun.Errorf("status lacks %q:\n%s", want, status)
		}
	}
	if strings.Contains(status, "refresh failed") {
		testRun.Error("an unavailable firewall is not a failure")
	}
}

func TestADisabledFirewallIsNeverReadAndSaysSo(testRun *testing.T) {
	model := newTestModel(testRun)
	updated, command := updateWith(testRun, model, RefreshResultMessage{Report: sampleReport(), At: testRefreshTime})
	if command != nil || updated.firewallRefreshing || !updated.firewallRequestedAt.IsZero() {
		testRun.Fatal("a disabled firewall must never be read")
	}
	updated, _ = updateWith(testRun, typeKeys(testRun, updated, "r"), tickMessage{})
	if updated.firewallRefreshing || updated.hasFirewall {
		testRun.Fatal("neither r nor a tick may read a disabled firewall")
	}
	for _, inner := range []InnerTab{0, 1} {
		rendered := showView(testRun, updated, moduleView{ModuleFirewall, inner}).View()
		if !strings.Contains(rendered, "Firewall checks are disabled (started with --no-firewall)") {
			testRun.Errorf("inner %d must say it is disabled:\n%s", inner, rendered)
		}
	}
	ports := typeKeys(testRun, showView(testRun, updated, moduleView{ModuleFirewall, 1}), "s", "/")
	if ports.firewallSortMode != FirewallSortByVerdict || ports.viewState != ViewTable {
		testRun.Fatal("sort and filter keys must be ignored while the firewall is disabled")
	}
	footer := strings.Join(ports.footerHintTexts(testWidth), footerHintJoin)
	if strings.Contains(footer, "s sort") || strings.Contains(footer, "/ filter") {
		testRun.Fatalf("disabled firewall must hide its footer hints: %q", footer)
	}
}

func TestFirewallCommandReportsAFailureAsAMessage(testRun *testing.T) {
	evaluator := &fakeFirewall{err: errFirewallDown}
	clock := &testClock{now: testRefreshTime}
	message, isFirewall := firewallCommand(evaluator, sampleReport().Ports, clock.read)().(FirewallResultMessage)
	if !isFirewall || message.Err == nil || evaluator.calls != 1 || !message.At.Equal(clock.now) {
		testRun.Fatalf("message=%+v calls=%d", message, evaluator.calls)
	}
}

func TestFirewallModuleHasStatusAndPortsTabs(testRun *testing.T) {
	model := showModule(testRun, evaluatedModel(testRun, sampleFirewallReport()), ModuleFirewall)
	if model.activeTab != TabFirewall {
		testRun.Fatalf("the module opens on Status, got %v", model.activeTab)
	}
	specs := ModuleFirewall.innerTabs()
	if len(specs) != 2 || specs[0].title != "Status" || specs[1].title != "Ports" || specs[1].view != TabFirewallPorts {
		testRun.Fatalf("inner tabs = %+v", specs)
	}
	if got := typeKeys(testRun, model, "2").activeTab; got != TabFirewallPorts {
		testRun.Fatalf("2 opens Ports, got %v", got)
	}
	if got := typeKeys(testRun, model, "]").activeTab; got != TabFirewallPorts {
		testRun.Fatalf("] opens Ports, got %v", got)
	}
	if got := typeKeys(testRun, model, "3").activeTab; got != TabFirewall {
		testRun.Fatalf("3 does not exist, got %v", got)
	}
}

func TestFirewallPortsAreSortedWithTheMostWorryingFirst(testRun *testing.T) {
	model := showView(testRun, evaluatedModel(testRun, sampleFirewallReport()), moduleView{ModuleFirewall, 1})
	// allowed 8080, restricted 3306, unknown 53, docker 443, blocked 22
	if got := visiblePorts(model); !equalInts(got, []int{8080, 3306, 53, 443, 22}) {
		testRun.Fatalf("verdict order = %v", got)
	}
	model = typeKeys(testRun, model, "s")
	if model.firewallSortMode != FirewallSortByPort || !equalInts(visiblePorts(model), []int{22, 53, 443, 3306, 8080}) {
		testRun.Fatalf("port order = %v (%v)", visiblePorts(model), model.firewallSortMode)
	}
	model = typeKeys(testRun, model, "s")
	if model.firewallSortMode != FirewallSortByProcess || !equalInts(visiblePorts(model), []int{53, 3306, 443, 22, 8080}) {
		testRun.Fatalf("process order = %v (%v)", visiblePorts(model), model.firewallSortMode)
	}
	model = typeKeys(testRun, model, "s")
	if model.firewallSortMode != FirewallSortByVerdict || model.sortMode != SortBySeverity {
		testRun.Fatal("s must wrap and must not touch the port sort mode")
	}
}

func TestFirewallPortsFilterByTextAndKeepTheirOwnFilter(testRun *testing.T) {
	model := showView(testRun, evaluatedModel(testRun, sampleFirewallReport()), moduleView{ModuleFirewall, 1})
	model = typeKeys(testRun, model, "/", "d", "r", "o", "p", "enter")
	if got := visiblePorts(model); !equalInts(got, []int{22}) {
		testRun.Fatalf("filter drop = %v", got)
	}
	if model.filterText != "" {
		testRun.Fatal("the port filter must stay untouched")
	}
	if !strings.Contains(model.View(), `filter:"drop"`) {
		testRun.Fatalf("the title must show the filter:\n%s", model.View())
	}
	byVerdict := typeKeys(testRun, showView(testRun, model, moduleView{ModuleFirewall, 1}), "/", "esc")
	if byVerdict.viewState != ViewTable {
		testRun.Fatal("esc closes the prompt")
	}
	ports := showModule(testRun, model, ModulePorts)
	if len(ports.visibleRows) != len(sampleReportWithQuietPort().Findings) {
		testRun.Fatalf("the Ports module must not be filtered by the firewall filter: %d rows", len(ports.visibleRows))
	}
}

func TestFirewallFooterHintsFollowTheView(testRun *testing.T) {
	model := evaluatedModel(testRun, sampleFirewallReport())
	status := strings.Join(showView(testRun, model, moduleView{ModuleFirewall, 0}).footerHintTexts(testWidth), footerHintJoin)
	for _, hidden := range []string{"s sort", "/ filter", "j/k", "f min-sev"} {
		if strings.Contains(status, hidden) {
			testRun.Errorf("Status must hide %q: %q", hidden, status)
		}
	}
	ports := strings.Join(showView(testRun, model, moduleView{ModuleFirewall, 1}).footerHintTexts(testWidth), footerHintJoin)
	for _, want := range []string{"s sort", "/ filter", "j/k"} {
		if !strings.Contains(ports, want) {
			testRun.Errorf("Ports lacks %q: %q", want, ports)
		}
	}
	if strings.Contains(ports, "f min-sev") {
		testRun.Errorf("Ports has no severity filter: %q", ports)
	}
}

func TestFirewallStatusKeysThatDoNotApplyAreIgnored(testRun *testing.T) {
	model := showView(testRun, evaluatedModel(testRun, sampleFirewallReport()), moduleView{ModuleFirewall, 0})
	model = typeKeys(testRun, model, "s", "f", "/", "j", "G", "l", "i")
	if model.firewallSortMode != FirewallSortByVerdict || model.viewState != ViewTable || model.cursor != 0 || !model.hideLoopback || model.inboundOnly {
		testRun.Fatal("Status must ignore sort, filter, cursor and connection keys")
	}
}

func TestFirewallPortsDetailsShowTheFullReason(testRun *testing.T) {
	model := showView(testRun, evaluatedModel(testRun, sampleFirewallReport()), moduleView{ModuleFirewall, 1})
	rendered := model.View()
	for _, want := range []string{"Details", "Reason: INPUT accepts new tcp/8080 from any source", "allowed"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("details lack %q:\n%s", want, rendered)
		}
	}
	model = typeKeys(testRun, model, "j")
	if !strings.Contains(model.View(), "Reason: INPUT accepts tcp/3306 only from 203.0.113.0/24") {
		testRun.Fatalf("the details must follow the cursor:\n%s", model.View())
	}
}

func TestPortsDetailsShowTheFirewallLine(testRun *testing.T) {
	model := evaluatedModel(testRun, sampleFirewallReport())
	// The sample findings are sorted by severity: the first row is the critical 3306.
	rendered := model.View()
	if !strings.Contains(rendered, "Firewall: restricted - INPUT accepts tcp/3306 only from 203.0.113.0/24") {
		testRun.Fatalf("details lack the firewall line:\n%s", rendered)
	}
	allPorts := typeKeys(testRun, model, "2")
	if !strings.Contains(allPorts.View(), "Firewall: ") {
		testRun.Fatalf("All ports details lack the firewall line:\n%s", allPorts.View())
	}
}

func TestPortsDetailsShowADashWhenTheFirewallWasNotEvaluated(testRun *testing.T) {
	disabled := newTestModel(testRun)
	if !strings.Contains(disabled.View(), "Firewall: -") {
		testRun.Fatalf("disabled firewall must show a dash:\n%s", disabled.View())
	}
	clock := &testClock{now: testRefreshTime}
	waiting, _ := firewallTestModel(testRun, &fakeFirewall{}, clock)
	if !strings.Contains(waiting.View(), "Firewall: -") {
		testRun.Fatalf("not evaluated yet must show a dash:\n%s", waiting.View())
	}
}

func TestOverviewShowsTheFirewallLine(testRun *testing.T) {
	evaluated := showModule(testRun, evaluatedModel(testRun, sampleFirewallReport()), ModuleOverview).View()
	if !strings.Contains(evaluated, "Firewall: 1 allowed, 1 restricted, 1 blocked, 1 unknown") {
		testRun.Errorf("overview lacks the counters:\n%s", evaluated)
	}
	unavailable := showModule(testRun, evaluatedModel(testRun, unavailableFirewallReport()), ModuleOverview).View()
	if !strings.Contains(unavailable, "Firewall: not available (needs root to read firewall rules)") {
		testRun.Errorf("overview must say the firewall is not available:\n%s", unavailable)
	}
	disabled := showModule(testRun, newTestModel(testRun), ModuleOverview).View()
	if !strings.Contains(disabled, "Firewall: checks disabled") {
		testRun.Errorf("overview must say it is disabled:\n%s", disabled)
	}
	clock := &testClock{now: testRefreshTime}
	waiting, _ := firewallTestModel(testRun, &fakeFirewall{}, clock)
	if !strings.Contains(showModule(testRun, waiting, ModuleOverview).View(), "Firewall: reading rules...") {
		testRun.Error("overview must say it is reading")
	}
}

func TestSidebarBadgeOfTheFirewallModule(testRun *testing.T) {
	clock := &testClock{now: testRefreshTime}
	waiting, _ := firewallTestModel(testRun, &fakeFirewall{}, clock)
	cases := []struct {
		name  string
		model Model
		want  sidebarBadge
	}{
		{"disabled", newTestModel(testRun), sidebarBadge{text: "off"}},
		{"not evaluated yet", waiting, sidebarBadge{}},
		{"unavailable", evaluatedModel(testRun, unavailableFirewallReport()), sidebarBadge{text: "n/a"}},
		{"allowed and restricted ports", evaluatedModel(testRun, sampleFirewallReport()), sidebarBadge{text: "2"}},
	}
	for _, testCase := range cases {
		if got := testCase.model.sidebarBadgeFor(ModuleFirewall); got != testCase.want {
			testRun.Errorf("%s: badge = %+v, want %+v", testCase.name, got, testCase.want)
		}
	}
	rendered := strings.Join(evaluatedModel(testRun, sampleFirewallReport()).sidebarPanel(evaluatedModel(testRun, sampleFirewallReport()).computeLayout()), "\n")
	if !strings.Contains(rendered, "Firewall") {
		testRun.Fatal("the sidebar must list the Firewall module")
	}
}

func TestFirewallSidebarEntryIsNoLongerAPlaceholder(testRun *testing.T) {
	if ModuleFirewall.isPlaceholder() || !ModuleAlerts.isPlaceholder() || !ModuleSettings.isPlaceholder() {
		testRun.Fatal("only Alerts and Settings are placeholders now")
	}
}

func TestFirewallStatusExplainsEveryVerdictAndTheInformationOnlyRule(testRun *testing.T) {
	model := showModule(testRun, evaluatedModel(testRun, sampleFirewallReport()), ModuleFirewall)
	model = sizedModel(testRun, model, 140, 50)
	rendered := model.View()
	for _, want := range []string{"Backend: iptables-save", "Available: yes", "Last evaluated: 03:04:05", "IPv6 rules were not readable",
		"ALLOWED 1", "RESTRICTED 1", "BLOCKED 1", "UNKNOWN 1", "DOCKER 1", "NOT-EXPOSED 0",
		"What the verdicts mean", "allowed: ", "restricted: ", "blocked: ", "unknown: ", "not-exposed: ", "docker: ",
		"never change severities"} {
		if !strings.Contains(rendered, want) {
			testRun.Errorf("status lacks %q:\n%s", want, rendered)
		}
	}
}

func TestFirewallColumnDropsBeforeAnyOtherPortColumn(testRun *testing.T) {
	spec := tableSpecFor(TabFindings)
	var firewallPriority, sourcePriority int
	for _, candidate := range spec.fixed {
		switch candidate.heading {
		case headingFirewallShort:
			firewallPriority = candidate.priority
		case headingSource:
			sourcePriority = candidate.priority
		}
	}
	if firewallPriority == 0 || firewallPriority <= sourcePriority {
		testRun.Fatalf("FW priority %d must be dropped before SOURCE %d", firewallPriority, sourcePriority)
	}
	if tableSpecFor(TabAllPorts).fixed[len(spec.fixed)-1].heading != headingFirewallShort {
		testRun.Fatal("All ports must carry the FW column too")
	}
	for contentWidth := 20; contentWidth <= 220; contentWidth++ {
		shown, _ := spec.layoutColumns(contentWidth)
		headings := make([]string, 0, len(shown))
		for _, candidate := range shown {
			headings = append(headings, candidate.heading)
		}
		joined := strings.Join(headings, ",")
		if strings.Contains(joined, headingFirewallShort) && !strings.Contains(joined, headingSource) {
			testRun.Fatalf("width %d keeps FW but drops SOURCE: %v", contentWidth, headings)
		}
	}
}

func TestFirewallColumnShowsOnWideTerminalsAndNeverCostsTheProcessColumn(testRun *testing.T) {
	model := evaluatedModel(testRun, sampleFirewallReport())
	for _, width := range []int{140, 120, 110, 100} {
		for _, inner := range []InnerTab{0, 1} {
			sized := showView(testRun, sizedModel(testRun, model, width, 40), moduleView{ModulePorts, inner})
			heading := sized.tableContent(sized.computeLayout())[0]
			if !strings.Contains(heading, headingProcessShort) {
				testRun.Errorf("width %d inner %d: PROCESS lost: %q", width, inner, heading)
			}
		}
	}
	wide := showView(testRun, sizedModel(testRun, model, 220, 40), moduleView{ModulePorts, 0})
	frame := wide.computeLayout()
	lines := wide.tableContent(frame)
	if !strings.Contains(lines[0], headingFirewallShort) || !strings.Contains(strings.Join(lines, "\n"), "restricted") {
		testRun.Fatalf("a wide terminal shows the FW column with verdict words:\n%s", strings.Join(lines, "\n"))
	}
}

func TestEveryFirewallViewStaysInsideEveryTerminalSize(testRun *testing.T) {
	models := map[string]Model{
		"data":        evaluatedModel(testRun, sampleFirewallReport()),
		"unavailable": evaluatedModel(testRun, unavailableFirewallReport()),
		"failed":      applyMessage(testRun, evaluatedModel(testRun, sampleFirewallReport()), FirewallResultMessage{Err: errFirewallDown, At: testRefreshTime}),
		"reading": func() Model {
			model, _ := firewallTestModel(testRun, &fakeFirewall{}, &testClock{now: testRefreshTime})
			return model
		}(),
		"disabled": newTestModel(testRun),
	}
	for _, base := range models {
		for _, size := range layoutSizes {
			sized := applyMessage(testRun, base, size)
			for _, target := range []moduleView{{ModuleFirewall, 0}, {ModuleFirewall, 1}, {ModulePorts, 0}, {ModulePorts, 1}, {ModuleOverview, 0}} {
				tabbed := showView(testRun, sized, target)
				for _, keys := range [][]string{nil, {"G"}, {"?"}, {"/"}, {"b"}, {"s", "j"}} {
					model := typeKeys(testRun, tabbed, keys...)
					assertRendersWithin(testRun, model, size.Width, size.Height)
				}
			}
		}
	}
}

func TestFirewallPortsHeadersStartAboveTheirCells(testRun *testing.T) {
	model := showView(testRun, evaluatedModel(testRun, sampleFirewallReport()), moduleView{ModuleFirewall, 1})
	for _, width := range []int{50, 70, 90, 110, 140} {
		sized := sizedModel(testRun, model, width, 40)
		frame := sized.computeLayout()
		lines := sized.tableContent(frame)
		spec := tableSpecFor(sized.activeTab)
		shown, _ := spec.layoutColumns(panelContentWidth(frame.contentWidth))
		offset := selectionMarkWidth + len(columnSeparator)
		for _, candidate := range shown {
			assertColumnAligned(testRun, lines, offset, candidate.width, candidate.heading, width, "firewall ports")
			offset += candidate.width + len(columnSeparator)
		}
		assertColumnAligned(testRun, lines, offset, frame.contentWidth, spec.flexHeading, width, "firewall ports")
	}
}

func TestFirewallPortsKeepTheVerdictAndPortColumnsOnNarrowTerminals(testRun *testing.T) {
	spec := tableSpecFor(TabFirewallPorts)
	shown, _ := spec.layoutColumns(10)
	headings := make([]string, 0, len(shown))
	for _, candidate := range shown {
		headings = append(headings, candidate.heading)
	}
	if strings.Join(headings, ",") != "VERDICT,PORT,PROTO" {
		testRun.Fatalf("mandatory columns = %v", headings)
	}
}
