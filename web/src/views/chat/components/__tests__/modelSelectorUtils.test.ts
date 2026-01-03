import { describe, expect, it } from 'vitest'
import { getInitialModelState } from '../modelSelectorUtils'

describe('getInitialModelState', () => {
  it('uses the default model without committing when session model is missing', () => {
    const result = getInitialModelState(undefined, 'model-a')
    expect(result).toEqual({ initialModel: 'model-a', shouldCommit: false })
  })

  it('commits when the session already has a model', () => {
    const result = getInitialModelState('model-a', 'model-b')
    expect(result).toEqual({ initialModel: 'model-a', shouldCommit: true })
  })
})
