/* global d3 */

const width = 960
const height = 500
const n = 20
const m = 200
const k = 10
const stack = d3.stack()
  .keys(d3.range(n))
  .order(d3.stackOrderNone) // stackOrderAscending  stackOrderDescending  stackOrderInsideOut
  .offset(d3.stackOffsetWiggle) // stackOffsetNone stackOffsetSilhouette

const svg = d3.select('body').append('svg')
  .attr('width', width)
  .attr('height', height)
  .append('g')
  // .attr('transform', 'translate(' + width / 2 + ',' + (height / 2) + ')')
// const svg = d3.select(DOM.svg(width, height))

const x = d3.scaleLinear()
  .domain([0, m - 1])
  .range([0, width])

const y = d3.scaleLinear()
  .range([height, 0])

const z = d3.interpolateCool

const area = d3.area()
  .x((d, i) => x(i))
  .y0(d => y(d[0]))
  .y1(d => y(d[1]))

function randomize () {
  const layers = stack(d3.transpose(Array.from({ length: n }, () => bumps(m, k))))
  y.domain([
    d3.min(layers, l => d3.min(l, d => d[0])),
    d3.max(layers, l => d3.max(l, d => d[1]))
  ])
  return layers
}

const path = svg.selectAll('path')
  .data(randomize)
  .enter().append('path')
  .attr('d', area)
  .attr('fill', () => z(Math.random()))

/* const interval = */ d3.interval(() => {
  path
    .data(randomize)
    .transition()
    .duration(1500)
    .attr('d', area)
}, 2500)

// related to observable?
// invalidation.then(() => interval.stop())

// Inspired by Lee Byron’s test data generator.
function bump (a, n) {
  const x = 1 / (0.1 + Math.random())
  const y = 2 * Math.random() - 0.5
  const z = 10 / (0.1 + Math.random())
  for (let i = 0; i < n; ++i) {
    const w = (i / n - y) * z
    a[i] += x * Math.exp(-w * w)
  }
}
function bumps (n, m) {
  const a = []
  for (let i = 0; i < n; ++i) a[i] = 0
  for (let i = 0; i < m; ++i) bump(a, n)
  return a
};
