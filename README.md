ROLE PERMISSIONS MATRIX
=============================

**Legend**  
- **V** = view  
- **E** = edit (non-critical fields: notes, photos, labor/material lines)  
- **T** = transition (change state)  
- **\*** = with constraints (see notes)


### Role Permission Table

| RESOURCE / ACTION                  | ADMIN          | ADJUSTER               | BODYMAN                 |
|-----------------------------------|----------------|-------------------------|--------------------------|
| Work order visibility (all)       | V (all)        | V (read-only)\*         | V (read-only)\*          |
| "My Assignments" visibility       | V              | V (if assigned)\*       | V (if assigned)          |
| Create intake (draft/WO)          | Yes            | Yes                     | Yes                      |
| Edit intake/basic info            | Yes            | Suggest via note         | Suggest via note          |
| Upload photos/docs                | Yes            | Yes                     | Yes                      |
| Set/Change assignee               | Yes            | No                      | No                       |
| Edit estimate lines (manual)      | Yes            | Suggest via note         | Yes (if assigned)         |
| Trigger AI estimate               | Yes            | View/Request            | Yes (if assigned)         |
| Send estimate to customer/insurer | Yes            | Suggest via note         | Suggest via note          |
| Set awaiting_party/reason         | Yes            | Yes                     | Yes (if assigned)         |
| Transition → **awaiting_info**    | Yes            | Suggest via note         | Yes (if assigned)\*       |
| Transition → **in_progress**      | Yes            | No                      | No                       |
| Transition → **insurance_denied** | Yes            | No                      | No                       |
| Transition → **completed**        | Yes            | No                      | No                       |
| Transition → **require_follow_up**| Yes            | Yes                     | Yes (if assigned)         |
| Switch to self-pay                | Yes            | Suggest via note         | Suggest via note          |
| Mark payment status               | Yes            | View                    | Suggest via note          |
| Parts status update               | Yes            | View                    | Yes (if assigned)         |
| Lock core fields on completed     | Yes            | No                      | No                       |
| Audit timeline view               | Yes            | Yes                     | Yes                      |

---

### Notes

- **Adjuster**: Positioned as a collaborator who provides documentation retrieval and damage assessment suggestions. They may initiate `request_more_info` and populate `awaiting_party` / `awaiting_reason`, but **cannot** push a work order into **in_progress** or **completed**.
  
- **Bodyman**: For assigned work orders, a bodyman may upload/edit labor & materials, trigger AI, and propose supplementary assessments. They may move a work order into **awaiting_info** (e.g., submitting estimate/materials), but **key business transitions** (in_progress, completed, insurance_denied) are performed by **Admin**.

- **Read-only visibility**: Same-shop team members can see work orders without overstepping permissions; this improves support and handoff.

