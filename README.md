=============================
ROLE PERMISSIONS MATRIX
=============================
Legend:
V = view
E = edit (non-critical fields: notes, photos, labor/material lines)
T = transition (change state)

- = with constraints (see notes)

---

## RESOURCE / ACTION ADMIN ADJUSTER BODYMAN

Work order visibility (all) V (all) V (read-only)_ V (read-only)_
"My Assignments" visibility V V (if assigned)_ V (if assigned)
Create intake (draft/WO) Yes Yes Yes
Edit intake/basic info Yes Suggest via note Suggest via note
Upload photos/docs Yes Yes Yes
Set/Change assignee Yes No No
Edit estimate lines (manual) Yes Suggest via note Yes (if assigned)
Trigger AI estimate Yes View/Request Yes (if assigned)
Send estimate to customer/insurer Yes Suggest via note Suggest via note
Set awaiting_party/reason Yes Yes Yes (if assigned)
Transition → awaiting_info Yes Suggest via note Yes (if assigned)_
Transition → in_progress Yes No No
Transition → insurance_denied Yes No No
Transition → completed Yes No No
Transition → require_follow_up Yes Yes Yes (if assigned)
Switch to self-pay Yes Suggest via note Suggest via note
Mark payment status Yes View Suggest via note
Parts status update Yes View Yes (if assigned)
Lock core fields on completed Yes No No
Audit timeline view Yes Yes Yes

---
