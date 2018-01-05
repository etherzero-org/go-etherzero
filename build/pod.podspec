Pod::Spec.new do |spec|
  spec.name         = 'Getz'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/ethzero/go-ethzero'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Ethzero Client'
  spec.source       = { :git => 'https://github.com/ethzero/go-ethzero.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Getz.framework'

	spec.prepare_command = <<-CMD
    curl https://getzstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Getz.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
