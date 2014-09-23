Then(/^no line of output exceeds (\d+) characters$/) do |n|
  assert { all_output.lines.map(&:length).max <= n.to_i }
end

Then(/^file "(.*?)" should contain an image$/) do |filename|
  in_current_dir do
    @image = ImageArchive.new(filename).image
  end
end

Then(/^the image's JSON should be like:$/) do |table|
  table.raw.each do |expr, value|
    assert { @image[expr] == value }
  end
end

Then(/^the image should add "\/?(.*?)"$/) do |path|
  assert { @image.include?(path) }
end

Then(/^the image should delete "\/?(.*?)"$/) do |path|
  assert { @image.delete?(path) }
end

Then(/^the image should not add "\/?(.*?)"$/) do |path|
  deny { @image.include?(path) }
end

Then(/^the image should not delete "\/?(.*?)"$/) do |path|
  deny { @image.delete?(path) }
end
