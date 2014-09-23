require 'json'
require 'set'
require 'zlib'

require 'archive/tar/minitar'
require 'jsonpath'
require 'wrong'

class ImageArchive
  extend Forwardable
  include Wrong::Assert
  include Enumerable

  Image = Struct.new('Image', :version, :json, :layer) do
    include Wrong::Assert
    extend Forwardable
    include Enumerable
    def_delegator :layer, :each
    def_delegator :layer, :size

    def id
      json['id']
    end

    def to_s
      "#<#{self.class} #{id}>"
    end

    # Return first and only result of JSONPath. Raise exception if
    # more than one. Return nil if no result.
    def [](expr)
      if expr.is_a? Symbol
        super
      else
        elts = jsonpath(expr).first(2)
        assert { elts.length < 2 }
        elts.first
      end
    end

    def jsonpath(expr)
      JsonPath.new(expr)[self.json]
    end

    def include?(path)
      layer.include?(path)
    end

    def delete?(path)
      dn = File.dirname(path)
      bn = ".wh.#{File.basename(path)}"
      include?(dn == '.' ? bn : File.join(dn, bn))
    end
  end
  attr_reader :images

  def_delegator '@images', :[]
  def_delegator '@images', :size
  def_delegator '@images.values', :each

  def initialize(io)
    if io.is_a? String
      gzipped = io.end_with?('gz')
      io = File.open(io, 'rb')
      io = Zlib::GzipReader.new(io) if gzipped
    end
    @images = {}
    directories = Set[]

    tar = Archive::Tar::Minitar::Reader.new(io)
    begin
      tar.each do |entry|
        next if %w'./ .'.include? entry.full_name
        assert { entry.full_name.count('/') == 1 }

        if entry.directory?
          directories << entry.full_name.sub(/\/$/, '')
          next
        end
        img = images[File.dirname(entry.full_name)] ||= Image[]
        case File.basename(entry.full_name)
        when 'VERSION'
          img.version = entry.read
        when 'json'
          img.json = JSON.load(entry.read)
        when 'layer.tar'
          img.layer = Set[]
          nfiles = 0
          ltar = Archive::Tar::Minitar::Reader.new(entry)
          begin
            ltar.each do |lentry|
              nfiles += 1
              img.layer << lentry.full_name
            end
          ensure
            ltar.close
          end
          assert { nfiles == img.layer.size }
        else
          deny { entry.full_name }
        end
      end
    ensure
      tar.close
    end

    # sanity check
    assert { size == directories.size }
    assert { Set.new(images.keys) == directories }
    images.each do |id, img|
      assert { img.id == id }
    end
    each do |img|
      assert { img.version == '1.0' }
    end
  end

  def image
    assert { size == 1 }
    return first
  end
end
