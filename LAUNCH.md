# Tempus - Launch Checklist

**Project**: Tempus - Neurodivergent-Friendly Calendar Tool
**Status**: Ready for Public Launch 🚀
**License**: MIT (Open Source)

---

## Pre-Launch Checklist

### ✅ Repository Setup (COMPLETED)

- [x] LICENSE file (MIT)
- [x] CODE_OF_CONDUCT.md (neurodivergent-friendly)
- [x] CONTRIBUTING.md (clear guidelines)
- [x] SECURITY.md (vulnerability disclosure)
- [x] Enhanced .gitignore (secrets protection)
- [x] CI/CD workflows (test, lint, build on 3 platforms)
- [x] Functional test suite in CI
- [x] Coverage checking (75% threshold)
- [x] Security scanning (weekly runs)
- [x] Release automation (6 platform binaries: Linux/macOS/Windows × AMD64/ARM64)
- [x] Docker support
- [x] golangci-lint config
- [x] Renovate bot configured
- [x] Git-flow workflows (sync-branches, promote-to-main)
- [x] Neurodivergent features documented (NEURODIVERGENT_FEATURES.md)
- [x] README with complete feature list
- [x] Repository created on GitHub
- [x] Description set: "ADHD-friendly ICS calendar generator (Go CLI)"
- [x] Branch protection enabled (main + develop)

### 📋 Before Going Public (TODO)

1. **GitHub Repository Settings**:
   - [ ] Add topics: `neurodivergent`, `adhd`, `autism`, `dyslexia`, `calendar`, `ics`, `golang`, `cli`, `rfc5545`, `productivity`, `accessibility`, `time-management`
   - [ ] Enable GitHub features:
     - [ ] Discussions
     - [ ] Wiki (optional)
     - [ ] Projects (optional)
   - [ ] Enable Security features (manually via UI):
     - [ ] Dependabot alerts (enable in Settings → Security)
     - [ ] Dependabot security updates
     - [ ] Code scanning (CodeQL)
     - [x] Secret scanning (already enabled)

3. **Create First Release**:
   - [ ] Tag version: `git tag v0.5.0 && git push origin v0.5.0`
   - [ ] Verify GitHub Actions creates release with binaries
   - [ ] Edit release notes to add highlights (see template below)

4. **External Services**:
   - [ ] Register at [Go Report Card](https://goreportcard.com/) - enter repo URL
   - [ ] (Optional) Set up [Codecov](https://codecov.io/) account for coverage badge
   - [ ] Verify repo appears on [pkg.go.dev](https://pkg.go.dev/) (automatic after ~1 hour)

5. **Verify No Secrets**:
   ```bash
   # Run locally before making public
   git log -p | grep -i "secret"
   git log -p | grep -i "password"
   git log -p | grep -i "token"
   git log -p | grep "@tempus"  # Check for test tokens
   ```

---

## Release Notes Template (v0.5.0)

```markdown
# Tempus v0.5.0 - Initial Public Release 🎉

**Tempus** is a neurodivergent-friendly ICS calendar generator designed for people with ADHD, ASD, Dyslexia, and other cognitive differences.

## ✨ Highlights

### 🧠 Neurodivergent-Friendly Features
- **Time-only input**: Type `10:30` instead of full datetime
- **Human durations**: `45m`, `1h30m`, `90 minutes` all work
- **Multiple reminders**: Fight time blindness with countdown alarms
- **Alarm Profiles**: Reusable alarm presets (adhd-default, adhd-countdown, medication)
- **Smart Duration Defaults**: Auto-detects sensible durations based on event type
- **Auto-Emoji Support**: Visual category icons (💊 medication, 💼 work, 🏥 health)
- **Input Normalization**: Auto-fixes date/time formats
- **Smart Spell Checking**: Corrects common typos (customizable)
- **Conflict Detection**: Detects overlapping events
- **Overwhelm Prevention**: Warns when days exceed event limit
- **Dry-Run Validation**: Preview events before creating
- **Batch Template Generator**: Pre-filled templates for common scenarios
- **RRULE Helper**: Interactive wizard to build recurrence rules
- **Prep Time Auto-Addition**: Automatic preparation/transition buffers (ADHD time boxing)

### 📅 Built-in Templates
- **flight**: Flight bookings with boarding reminders
- **meeting**: Team meetings
- **holiday**: Vacation periods
- **medical**: Medical appointments with travel time
- **focus-block**: 90-minute deep work sessions
- **medication**: Medication reminders with triple alarms
- **appointment**: Appointments with travel time
- **transition**: 15-minute buffer periods
- **deadline**: Countdown reminders for deadlines

### 🔧 Technical Features
- RFC 5545 compliant ICS generation
- Smart timezone handling (different TZs for start/end)
- Batch operations (CSV/JSON/YAML)
- Multilingual (EN, ES, PT, GA)
- Import to any calendar app (Google Calendar, Apple Calendar, Outlook)

## 📦 Download

Binaries available for:
- **Linux** (AMD64, ARM64)
- **macOS** (Intel, Apple Silicon)
- **Windows** (AMD64, ARM64)

Download from [Releases](https://github.com/malpanez/tempus/releases/tag/v0.5.0)

## 📚 Documentation

- [Quick Start Guide](https://github.com/malpanez/tempus#quick-start)
- [Neurodivergent Features Guide](https://github.com/malpanez/tempus/blob/main/docs/NEURODIVERGENT_FEATURES.md)
- [Batch Examples](https://github.com/malpanez/tempus/blob/main/examples/README.md)

## 🙏 Acknowledgments

This project was built with [Claude Code](https://claude.com/claude-code) AI assistance, combining lived experience with ADHD and modern development tools to create something genuinely helpful for the neurodivergent community.

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

**Full Changelog**: https://github.com/malpanez/tempus/commits/v0.5.0
```

---

## Community Announcement Templates

### Reddit - r/ADHD

**Title**: "I built a CLI calendar tool for ADHD, ASD, and Dyslexia (free, open-source)"

**Body**:
```
Hey everyone! I'm neurodivergent and struggled with traditional calendar tools, so I built **Tempus** - a command-line calendar generator designed for ADHD, ASD, Dyslexia, and other cognitive differences.

**Why it's neurodivergent-friendly:**

🧠 **Reduces decision fatigue**:
- Type just the time (10:30) instead of full datetime
- Durations in plain English (45m, 1h30m, 90 minutes)
- Smart defaults so you don't have to think

⏰ **Fights time blindness**:
- Multiple countdown reminders (1 week, 3 days, 1 day, morning)
- "Transition time" templates (15min buffers between tasks)
- Medication reminders with 3 alarms (-10min, 0min, +5min)
- Prep time auto-addition (15min before meetings, based on ADHD research)

📱 **Practical features**:
- Focus block templates (90min deep work + prep reminder)
- Appointments with built-in travel time calculation
- Generates ICS files that work with Google/Apple/Outlook calendars
- Works offline, your data stays private
- Conflict detection and overwhelm prevention
- Dry-run mode to preview before creating

**It's free, open-source (MIT license), and works on Windows/Mac/Linux.**

GitHub: https://github.com/malpanez/tempus

I built this because I needed it, and I hope it helps others too. Feedback welcome!

(Built with AI assistance - transparent about the process because I wanted to ship something helpful quickly rather than spend months learning every detail of RFC 5545 😅)
```

---

### Hacker News - Show HN

**Title**: "Show HN: Tempus – Neurodivergent-friendly CLI for calendar events (ADHD/ASD/Dyslexia)"

**Body**:
```
Hi HN! I'm a neurodivergent developer who struggled with traditional calendar tools, so I built Tempus - a command-line ICS generator with a neurodivergent-friendly UX (ADHD, ASD, Dyslexia).

**What makes it neurodivergent-friendly:**

1. **Time-only input**: Type "10:30" instead of "2025-11-15 10:30:00" (auto-expands to today)
2. **Duration parsing**: "45m", "1h30m", "90 minutes", "1:15" all work
3. **Multiple reminders**: Templates with countdown alarms (1 week, 3 days, 1 day, morning)
4. **Smart defaults**: Minimize decision fatigue
5. **Prep time auto-addition**: Based on ADHD time boxing research (15min buffers)
6. **Conflict detection**: Warns about overlapping events
7. **Overwhelm prevention**: Warns when days have too many events

**ADHD-specific templates:**
- Medication reminders (3 alarms: -10min, 0min, +5min)
- Focus blocks (90min deep work + prep reminder)
- Appointments (with travel time calculation)
- Transition buffers (15min between tasks)
- Deadlines (escalating countdown reminders)

**Tech stack:**
- Go 1.24 with Cobra/Viper
- RFC 5545 compliant ICS generation
- Embedded IANA timezone database
- Universal calendar compatibility (no vendor lock-in)
- 78.8% test coverage
- CI/CD with multi-platform releases (Linux, macOS, Windows × AMD64/ARM64)

**Why CLI?**: Many neurodivergent individuals prefer keyboard-driven workflows (fewer distractions, faster input, scriptable, consistent interface).

The tool is MIT-licensed. I built it with Claude Code AI assistance to accelerate development while maintaining production quality.

GitHub: https://github.com/malpanez/tempus

I'd love feedback on the UX decisions, code architecture, or feature suggestions!
```

---

### Twitter/X

```
🚀 Just open-sourced Tempus - a CLI calendar tool for neurodivergent users (ADHD, ASD, Dyslexia)

✅ Time-only input (just type "10:30")
✅ Spell checking for dyslexia (auto-corrects typos)
✅ Multiple reminders to fight time blindness
✅ Focus block & medication templates
✅ Works with Google/Apple/Outlook calendars
✅ Conflict detection & overwhelm prevention
✅ 100% free & open-source

Built with @AnthropicAI Claude Code 🤖

https://github.com/malpanez/tempus

#neurodivergent #ADHD #autism #dyslexia #golang #CLI #opensource
```

---

## Launch Day Timeline

**Hour 0** (Morning):
- [ ] Push final changes to main
- [ ] Create release tag: `git tag v0.5.0 && git push origin v0.5.0`
- [ ] Verify GitHub Actions completes successfully
- [ ] Check binaries are downloadable

**Hour 1**:
- [ ] Post to r/ADHD (largest audience)
- [ ] Post to r/golang
- [ ] Post to r/commandline
- [ ] Tweet announcement

**Hour 2-3**:
- [ ] Submit to Hacker News "Show HN"
- [ ] Post to LinkedIn (optional)

**Hour 4-24**:
- [ ] Monitor comments/issues
- [ ] Respond to questions promptly
- [ ] Thank everyone for feedback
- [ ] Fix critical bugs if found

**Day 2-7**:
- [ ] Continue engagement
- [ ] Submit to awesome lists
- [ ] Plan v0.6.0 based on feedback

---

## Emergency Response Plan

### If negative feedback:
- Stay calm and professional
- Listen to criticism (there's often truth in it)
- Fix legitimate bugs quickly
- Clarify misunderstandings politely
- Update docs if confusion is common

### If security issue:
- Follow SECURITY.md process
- Patch immediately
- Release hotfix version
- Notify users via GitHub Security Advisory

### If overwhelmed:
- It's okay to slow down
- Update README: "Looking for maintainers"
- You helped people - that's success

---

## Success Metrics

### Short-term (1 month)
- [ ] 50+ GitHub stars
- [ ] 5+ issues opened
- [ ] 1+ external contributor
- [ ] 100+ downloads

### Medium-term (6 months)
- [ ] 200+ stars
- [ ] 10+ contributors
- [ ] Featured in 1+ "awesome" list
- [ ] 500+ downloads

### Long-term (1 year)
- [ ] 500+ stars
- [ ] 25+ contributors
- [ ] Package in Homebrew/Snap/AUR
- [ ] 2000+ downloads

---

## FAQ Responses (for community engagement)

### "Why not just use Google Calendar directly?"
"Google Calendar is great for viewing, but repetitive event creation is tedious. Tempus is for power users who want to batch-create events, script calendar generation, or work offline. The tool creates standard ICS files that work with ANY calendar app - no vendor lock-in. Plus, neurodivergent users often prefer CLI workflows (faster, fewer distractions, consistent interface)."

### "Is this really needed?"
"For neurotypical users, maybe not. But for neurodivergent individuals (ADHD, ASD, Dyslexia), traditional calendar tools have too much cognitive overhead. Tempus reduces decision fatigue with smart defaults, corrects spelling automatically, provides multiple reminders to fight time blindness, and uses visual aids for quick scanning - all critical neurodivergent challenges."

### "Why CLI and not GUI?"
"Many neurodivergent individuals prefer keyboard-driven workflows (fewer visual distractions, faster input, scriptable). That said, a web UI is on the roadmap for future versions!"

### "Built with AI - is that cheating?"
"No! AI-assisted development is the future. I designed the features based on lived neurodivergent experience, and used AI to accelerate implementation. The result is production-quality code that helps real people. That's what matters."

---

## Next Steps After Launch

1. **Monitor initial feedback** (first 48 hours)
   - Respond to all comments
   - Fix critical bugs immediately
   - Note feature requests for roadmap

2. **Create issues from feedback**
   - Label by priority
   - Tag good first issues
   - Assign milestones

3. **Engage with contributors**
   - Thank everyone
   - Provide clear feedback on PRs
   - Merge and credit contributions

4. **Iterate on documentation**
   - Add FAQ section
   - Create troubleshooting guide

5. **Plan v0.6.0 release**
   - Based on community feedback
   - Maintain neurodivergent-friendly focus

---

**YOU'VE GOT THIS! 🚀**

The project is production-ready, well-documented, and genuinely helpful. The neurodivergent community will appreciate your contribution.

Remember: Even if only 10 people use it, you've made their lives better. That's success.
