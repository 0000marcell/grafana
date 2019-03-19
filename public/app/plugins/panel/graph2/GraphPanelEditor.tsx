//// Libraries
import _ from 'lodash';
import React, { PureComponent } from 'react';

// Types
import { PanelEditorProps, Switch, LegendEditor, LegendOptions, StatID } from '@grafana/ui';
import { GraphOptions } from './types';

export class GraphPanelEditor extends PureComponent<PanelEditorProps<GraphOptions>> {
  onToggleLines = () => {
    this.props.onOptionsChange({ ...this.props.options, showLines: !this.props.options.showLines });
  };

  onToggleBars = () => {
    this.props.onOptionsChange({ ...this.props.options, showBars: !this.props.options.showBars });
  };

  onTogglePoints = () => {
    this.props.onOptionsChange({ ...this.props.options, showPoints: !this.props.options.showPoints });
  };

  onLegendOptionsChange = (options: LegendOptions) => {
    this.props.onOptionsChange({ ...this.props.options, legend: options });
  }

  render() {
    const { showBars, showPoints, showLines } = this.props.options;

    return (
      <div>
        <div className="section gf-form-group">
          <h5 className="section-heading">Draw Modes</h5>
          <Switch label="Lines" labelClass="width-5" checked={showLines} onChange={this.onToggleLines} />
          <Switch label="Bars" labelClass="width-5" checked={showBars} onChange={this.onToggleBars} />
          <Switch label="Points" labelClass="width-5" checked={showPoints} onChange={this.onTogglePoints} />
        </div>
        <div className="section gf-form-group">
          <h5 className="section-heading">Test Options</h5>
          <Switch label="Lines" labelClass="width-5" checked={showLines} onChange={this.onToggleLines} />
          <Switch label="Bars" labelClass="width-5" checked={showBars} onChange={this.onToggleBars} />
          <Switch label="Points" labelClass="width-5" checked={showPoints} onChange={this.onTogglePoints} />
        </div>
        <div className="section gf-form-group">
          <LegendEditor
            stats={[StatID.min, StatID.max]}
            options={this.props.options.legend}
            onChange={this.onLegendOptionsChange}
          />
        </div>
      </div>
    );
  }
}
