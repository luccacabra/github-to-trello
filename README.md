# github-to-trello
Customizable GitHub to Trello synchronisation task

## Why
* 


## SetUp

1. Create card with unique name and master set of labels applied

## Develop

## Deploy

## Test


## Config
* Currently only supports label assignment by name (not color)
* Does not support pagination (only checks first 100 issues, prs, comments etc)
* No label/list assignment inheritance
* Only syncs to a single board
### Precedence
`base` < `issue_type#type` < (`card_action#action` | `user_relation#relation`)
### all
```yaml
github_org_name:

trello_board_name:
trello_label_card_name:
trello_labels:
  - color:
  - name:
trello_lists:
  -
  
sync_actions: !sync_actions
```

### sync actions (open | update | close)
```yaml
<action>:
    labels:
      - color:
      - name:
    lists:
      -
    issue_types:
      issue: !issue
      pull_request: !pull_request
```

### issue
```yaml
labels:
  - color:
  - name:
lists:
  -
user_relationship: !issue_user_relationship
```

### pull request
```yaml
labels:
  - color:
  - name:
lists:
  -
user_relationship:
    !issue_user_relationship
    review_requested:
      team:
        labels:
          - color:
          - name:
        lists:
          -
      user:
        labels:
          - color:
          - name:
        lists:
          -
```

### issue user relationship
```yaml
assignee:
  labels:
    - color:
    - name:
mentioned:
  labels:
    - color:
    - name:
  lists:
    -

```