import React, { PureComponent } from 'react';
import { connect } from 'react-redux';
import { setUsersSearchQuery } from './state/actions';
import { getInviteesCount, getUsersSearchQuery } from './state/selectors';
import { FilterInput } from 'app/core/components/FilterInput/FilterInput';
import { ToggleButton, ToggleButtonGroup } from '@grafana/ui';

export interface Props {
  searchQuery: string;
  setUsersSearchQuery: typeof setUsersSearchQuery;
  onShowInvites: () => void;
  pendingInvitesCount: number;
  canInvite: boolean;
  showInvites: boolean;
  externalUserMngLinkUrl: string;
  externalUserMngLinkName: string;
}

export class UsersActionBar extends PureComponent<Props> {
  render() {
    const {
      canInvite,
      externalUserMngLinkName,
      externalUserMngLinkUrl,
      searchQuery,
      pendingInvitesCount,
      setUsersSearchQuery,
      onShowInvites,
      showInvites,
    } = this.props;

    return (
      <div className="page-action-bar">
        <div className="gf-form gf-form--grow">
          <FilterInput
            labelClassName="gf-form--has-input-icon"
            inputClassName="gf-form-input width-20"
            value={searchQuery}
            onChange={setUsersSearchQuery}
            placeholder="Filter by name or type"
          />
          {pendingInvitesCount > 0 && (
            <div style={{ marginLeft: '1rem' }}>
              <ToggleButtonGroup>
                <ToggleButton selected={!showInvites} key="users" onChange={onShowInvites}>
                  Users
                </ToggleButton>
                <ToggleButton selected={showInvites} onChange={onShowInvites} key="pending-invites">
                  Pending Invites ({pendingInvitesCount})
                </ToggleButton>
              </ToggleButtonGroup>
            </div>
          )}
          <div className="page-action-bar__spacer" />
          {canInvite && (
            <a className="btn btn-primary" href="org/users/invite">
              <span>Invite</span>
            </a>
          )}
          {externalUserMngLinkUrl && (
            <a className="btn btn-primary" href={externalUserMngLinkUrl} target="_blank">
              <i className="fa fa-external-link-square" /> {externalUserMngLinkName}
            </a>
          )}
        </div>
      </div>
    );
  }
}

function mapStateToProps(state: any) {
  return {
    searchQuery: getUsersSearchQuery(state.users),
    pendingInvitesCount: getInviteesCount(state.users),
    externalUserMngLinkName: state.users.externalUserMngLinkName,
    externalUserMngLinkUrl: state.users.externalUserMngLinkUrl,
    canInvite: state.users.canInvite,
  };
}

const mapDispatchToProps = {
  setUsersSearchQuery,
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(UsersActionBar);
